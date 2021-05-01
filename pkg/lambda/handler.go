package lambda

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	wh "github.com/suzuki-shunsuke/lambuild/pkg/webhook"
	"github.com/suzuki-shunsuke/matchfile-parser/matchfile"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Handler struct {
	Config          config.Config
	Secret          Secret
	SecretID        string
	SecretVersionID string
	Region          string
	GitHub          *github.Client
	CodeBuild       *codebuild.CodeBuild
	MatchfileParser *matchfile.Parser
}

func (handler *Handler) getRepo(repoName string) (config.Repository, bool) {
	for _, repo := range handler.Config.Repositories {
		if repo.Name == repoName {
			return repo, true
		}
	}
	return config.Repository{}, false
}

func (handler *Handler) matchHook(webhook wh.Webhook, event wh.Event, repo config.Repository, hook config.Hook) (bool, error) {
	if len(hook.Event) != 0 {
		eventMatched := false
		for _, ev := range hook.Event {
			if ev == webhook.Headers.Event {
				eventMatched = true
				break
			}
		}
		if !eventMatched {
			return false, nil
		}
	}

	if hook.Ref != "" {
		f, err := handler.MatchfileParser.Match([]string{event.Ref}, hook.RefConditions)
		if err != nil {
			return false, err
		}
		if !f {
			return false, nil
		}
	}

	if hook.BaseRef != "" {
		f, err := handler.MatchfileParser.Match([]string{event.BaseRef}, hook.BaseRefConditions)
		if err != nil {
			return false, err
		}
		if !f {
			return false, nil
		}
	}

	if event.PRAuthor != "" && hook.Author != "" {
		f, err := handler.MatchfileParser.Match([]string{event.PRAuthor}, hook.AuthorConditions)
		if err != nil {
			return false, err
		}
		if !f {
			return false, nil
		}
	}

	if webhook.Headers.Event == "pull_request" && hook.Label != "" {
		f, err := handler.MatchfileParser.Match(event.Labels, hook.LabelConditions)
		if err != nil {
			return false, err
		}
		if !(hook.NoLabel && len(event.Labels) == 0 && !f) {
			if !f {
				return false, nil
			}
		}
	}

	return true, nil
}

func (handler *Handler) getHook(webhook wh.Webhook, event wh.Event, repo config.Repository) (config.Hook, bool, error) {
	for _, hook := range repo.Hooks {
		f, err := handler.matchHook(webhook, event, repo, hook)
		if err != nil {
			return config.Hook{}, false, err
		}
		if f {
			return hook, true, nil
		}
	}
	return config.Hook{}, false, nil
}

func (handler *Handler) handleEvent(ctx context.Context, webhook wh.Webhook, event wh.Event) error {
	event.RepoOwner = strings.Split(event.RepoFullName, "/")[0]
	logE := logrus.WithFields(logrus.Fields{
		"repo": event.RepoFullName,
		"ref":  event.Ref,
	})

	repo, f := handler.getRepo(event.RepoFullName)
	if !f {
		logE.Debug("no repo matches")
		return nil
	}

	hook, f, err := handler.getHook(webhook, event, repo)
	if err != nil {
		return err
	}
	if !f {
		logE.Debug("no hook matches")
		return nil
	}
	logE = logE.WithFields(logrus.Fields{
		"config": hook.Config,
		"event":  hook.Event,
	})

	if event.HeadCommitMessage == "" {
		commit, _, err := handler.GitHub.Git.GetCommit(ctx, event.RepoOwner, event.RepoName, event.SHA)
		if err != nil {
			return fmt.Errorf("get a commit: %w", err)
		}
		event.HeadCommitMessage = commit.GetMessage()
	}

	file, _, _, err := handler.GitHub.Repositories.GetContents(ctx, event.RepoOwner, event.RepoName, hook.Config, &github.RepositoryContentGetOptions{Ref: event.Ref})
	if err != nil {
		logE.WithError(err).Debug("no content is found")
		return nil
	}
	content, err := file.GetContent()
	if err != nil {
		return fmt.Errorf("get a content: %w", err)
	}

	buildspec := bspec.Buildspec{}
	if err := yaml.Unmarshal([]byte(content), &buildspec); err != nil {
		return fmt.Errorf("unmarshal a buildspec: %w", err)
	}

	envs := []*codebuild.EnvironmentVariable{
		{
			Name:  aws.String("LAMBUILD_WEBHOOK_BODY"),
			Value: aws.String(webhook.Body),
		},
		{
			Name:  aws.String("LAMBUILD_WEBHOOK_EVENT"),
			Value: aws.String(webhook.Headers.Event),
		},
		{
			Name:  aws.String("LAMBUILD_WEBHOOK_DELIVERY"),
			Value: aws.String(webhook.Headers.Delivery),
		},
		{
			Name:  aws.String("LAMBUILD_HEAD_COMMIT_MSG"),
			Value: aws.String(event.HeadCommitMessage),
		},
	}

	if event.PRNum == 0 {
		prNum, err := getPRNumber(ctx, event.RepoOwner, event.RepoName, event.SHA, handler.GitHub)
		if err != nil {
			return err
		}
		event.PRNum = prNum
	}

	if event.PRNum != 0 {
		if event.PRAuthor == "" {
			pr, _, err := handler.GitHub.PullRequests.Get(ctx, event.RepoOwner, event.RepoName, event.PRNum)
			if err != nil {
				return fmt.Errorf("get a pull request: %w", err)
			}
			handler.setPullRequest(&event, pr)
		}

		files, _, err := getPRFiles(ctx, handler.GitHub, event.RepoOwner, event.RepoName, event.PRNum, event.ChangedFiles)
		if err != nil {
			return fmt.Errorf("get pull request files: %w", err)
		}
		event.ChangedFileNames = extractPRFileNames(files)

		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String("LAMBUILD_PR_NUMBER"),
			Value: aws.String(strconv.Itoa(event.PRNum)),
		}, &codebuild.EnvironmentVariable{
			Name:  aws.String("LAMBUILD_PR_AUTHOR"),
			Value: aws.String(event.PRAuthor),
		}, &codebuild.EnvironmentVariable{
			Name:  aws.String("LAMBUILD_PR_BASE_REF"),
			Value: aws.String(event.BaseRef),
		}, &codebuild.EnvironmentVariable{
			Name:  aws.String("LAMBUILD_PR_HEAD_REF"),
			Value: aws.String(event.HeadRef),
		})
	}

	if len(buildspec.Batch.BuildGraph) != 0 {
		return handler.handleGraph(ctx, logE, event, webhook, buildspec, repo, envs)
	}
	if len(buildspec.Batch.BuildList) != 0 {
		return handler.handleList(ctx, logE, event, webhook, buildspec, repo, envs)
	}
	if !buildspec.Batch.BuildMatrix.Empty() {
		return handler.handleMatrix(ctx, logE, event, webhook, buildspec, repo, envs)
	}

	return nil
}

func (handler *Handler) setPullRequest(event *wh.Event, pr *github.PullRequest) {
	event.Ref = pr.GetHead().GetRef()
	event.Labels = extractLabelNames(pr.Labels)
	event.PRAuthor = pr.GetUser().GetLogin()
	event.ChangedFiles = pr.GetChangedFiles()
	event.BaseRef = pr.GetBase().GetRef()
	event.HeadRef = pr.GetHead().GetRef()
}

func (handler *Handler) setRepo(event *wh.Event, repo *github.Repository) {
	event.RepoFullName = repo.GetFullName()
	event.RepoName = repo.GetName()
}

func (handler *Handler) handlePushEvent(ctx context.Context, webhook wh.Webhook, event *github.PushEvent) error {
	repo := event.GetRepo()
	ev := wh.Event{
		RepoFullName:      repo.GetFullName(),
		RepoName:          repo.GetName(),
		Ref:               event.GetRef(),
		SHA:               event.GetAfter(),
		HeadCommitMessage: event.GetHeadCommit().GetMessage(),
	}
	return handler.handleEvent(ctx, webhook, ev)
}

func (handler *Handler) handlePREvent(ctx context.Context, webhook wh.Webhook, event *github.PullRequestEvent) error {
	repo := event.GetRepo()
	pr := event.GetPullRequest()
	ev := wh.Event{
		SHA:   pr.GetHead().GetSHA(),
		PRNum: pr.GetNumber(),
	}
	handler.setRepo(&ev, repo)
	handler.setPullRequest(&ev, pr)
	return handler.handleEvent(ctx, webhook, ev)
}

func (handler *Handler) Do(ctx context.Context, webhook wh.Webhook) error {
	if err := github.ValidateSignature(webhook.Headers.Signature, []byte(webhook.Body), []byte(handler.Secret.WebhookSecret)); err != nil {
		// TODO return 400
		fmt.Println(err)
		return nil
	}
	body, err := github.ParseWebHook(webhook.Headers.Event, []byte(webhook.Body))
	if err != nil {
		return fmt.Errorf("parse a webhook payload: %w", err)
	}

	switch webhook.Headers.Event {
	case "push":
		return handler.handlePushEvent(ctx, webhook, body.(*github.PushEvent))
	case "pull_request":
		return handler.handlePREvent(ctx, webhook, body.(*github.PullRequestEvent))
	default:
		return nil
	}
}

func (handler *Handler) Init(ctx context.Context) error {
	cfg := config.Config{}
	configRaw := os.Getenv("CONFIG")
	if configRaw == "" {
		return errors.New("the environment variable 'CONFIG' is required")
	}
	if err := yaml.Unmarshal([]byte(configRaw), &cfg); err != nil {
		return fmt.Errorf("parse the environment variable 'CONFIG' as YAML: %w", err)
	}

	if len(cfg.Repositories) == 0 {
		return errors.New(`the configuration 'repositories' is required`)
	}
	for _, repo := range cfg.Repositories {
		if repo.Name == "" {
			return errors.New(`the repository 'name' is required`)
		}
		if repo.CodeBuild.ProjectName == "" {
			return fmt.Errorf(`'project-name' is required (repo: %s)`, repo.Name)
		}
	}

	handler.Config = cfg
	handler.Region = os.Getenv("REGION")
	if handler.Region == "" {
		return errors.New("the environment variable 'REGION' is required")
	}
	sess := session.Must(session.NewSession())
	secretNameGitHubToken := os.Getenv("SSM_PARAMETER_NAME_GITHUB_TOKEN")
	if secretNameGitHubToken == "" {
		return errors.New("the environment variable 'SSM_PARAMETER_NAME_GITHUB_TOKEN' is required")
	}
	secretNameWebhookSecret := os.Getenv("SSM_PARAMETER_NAME_WEBHOOK_SECRET")
	if secretNameWebhookSecret != "" {
		return errors.New("the environment variable 'SSM_PARAMETER_NAME_WEBHOOK_SECRET' is required")
	}

	if err := handler.readSecretFromSSM(ctx, sess, secretNameGitHubToken, secretNameWebhookSecret); err != nil {
		return err
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"log_level": logLevel,
			}).WithError(err).Error("the log level is invalid")
		} else {
			logrus.SetLevel(lvl)
		}
	}

	handler.GitHub = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: handler.Secret.GitHubToken},
	)))
	handler.CodeBuild = codebuild.New(sess, aws.NewConfig().WithRegion(handler.Region))

	handler.MatchfileParser = matchfile.NewParser()
	for i, repo := range cfg.Repositories {
		for j, hook := range repo.Hooks {
			for _, ev := range hook.Event {
				if ev != "push" && ev != "pull_request" {
					return fmt.Errorf(`hook.event is invalid: %s`, ev)
				}
			}

			conditions, err := handler.MatchfileParser.ParseConditions(strings.Split(strings.TrimSpace(hook.Ref), "\n"))
			if err != nil {
				return err
			}
			hook.RefConditions = conditions

			conditions, err = handler.MatchfileParser.ParseConditions(strings.Split(strings.TrimSpace(hook.BaseRef), "\n"))
			if err != nil {
				return err
			}
			hook.BaseRefConditions = conditions

			conditions, err = handler.MatchfileParser.ParseConditions(strings.Split(strings.TrimSpace(hook.Label), "\n"))
			if err != nil {
				return err
			}
			hook.LabelConditions = conditions

			conditions, err = handler.MatchfileParser.ParseConditions(strings.Split(strings.TrimSpace(hook.Author), "\n"))
			if err != nil {
				return err
			}
			hook.AuthorConditions = conditions

			repo.Hooks[j] = hook
		}
		cfg.Repositories[i] = repo
	}

	return nil
}
