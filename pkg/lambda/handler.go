package lambda

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Handler struct {
	Config             config.Config
	Secret             Secret
	SecretID           string
	SecretVersionID    string
	Region             string
	BuildStatusContext *template.Template
	GitHub             *github.Client
	CodeBuild          *codebuild.CodeBuild
}

func (handler *Handler) getRepo(repoName string) (config.Repository, bool) {
	for _, repo := range handler.Config.Repositories {
		if repo.Name == repoName {
			return repo, true
		}
	}
	return config.Repository{}, false
}

func runExpr(prog *vm.Program, data Data) (interface{}, error) {
	return expr.Run(prog, setExprFuncs(map[string]interface{}{
		"event": data.Event,
		"pr":    data.PullRequest,
		"repo":  data.Repository,
		"sha":   data.SHA,
		"ref":   data.Ref,
	}))
}

func (handler *Handler) matchHook(data Data, repo config.Repository, hook config.Hook) (bool, error) {
	if hook.If == nil {
		return true, nil
	}
	f, err := runExpr(hook.If, data)
	if err != nil {
		return false, fmt.Errorf("evaluate an expression: %w", err)
	}
	return f.(bool), nil
}

func (handler *Handler) getHook(data Data, repo config.Repository) (config.Hook, bool, error) {
	for _, hook := range repo.Hooks {
		f, err := handler.matchHook(data, repo, hook)
		if err != nil {
			return config.Hook{}, false, err
		}
		if f {
			return hook, true, nil
		}
	}
	return config.Hook{}, false, nil
}

func (handler *Handler) handleEvent(ctx context.Context, data Data) error {
	data.Repository.Owner = strings.Split(data.Repository.FullName, "/")[0]
	logE := logrus.WithFields(logrus.Fields{
		"repo_full_name": data.Repository.FullName,
		"repo_owner":     data.Repository.Owner,
		"repo_name":      data.Repository.Name,
		"ref":            data.Ref,
	})

	repo, f := handler.getRepo(data.Repository.FullName)
	if !f {
		logE.Debug("no repo matches")
		return nil
	}

	hook, f, err := handler.getHook(data, repo)
	if err != nil {
		return err
	}
	if !f {
		logE.Debug("no hook matches")
		return nil
	}
	logE = logE.WithFields(logrus.Fields{
		"config": hook.Config,
	})

	if data.HeadCommitMessage == "" {
		commit, _, err := handler.GitHub.Git.GetCommit(ctx, data.Repository.Owner, data.Repository.Name, data.SHA)
		if err != nil {
			return fmt.Errorf("get a commit: %w", err)
		}
		data.HeadCommitMessage = commit.GetMessage()
	}

	file, _, _, err := handler.GitHub.Repositories.GetContents(ctx, data.Repository.Owner, data.Repository.Name, hook.Config, &github.RepositoryContentGetOptions{Ref: data.Ref})
	if err != nil {
		logE.WithError(err).Debug("no content is found")
		return nil
	}
	if file == nil {
		logE.Warn("downloaded file is nil")
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
			Value: aws.String(data.Event.Body),
		},
		{
			Name:  aws.String("LAMBUILD_WEBHOOK_EVENT"),
			Value: aws.String(data.Event.Headers.Event),
		},
		{
			Name:  aws.String("LAMBUILD_WEBHOOK_DELIVERY"),
			Value: aws.String(data.Event.Headers.Delivery),
		},
		{
			Name:  aws.String("LAMBUILD_HEAD_COMMIT_MSG"),
			Value: aws.String(data.HeadCommitMessage),
		},
	}

	prNum := 0
	if data.PullRequest.PullRequest == nil {
		n, err := getPRNumber(ctx, data.Repository.Owner, data.Repository.Name, data.SHA, handler.GitHub)
		if err != nil {
			return err
		}
		prNum = n
	} else {
		prNum = data.PullRequest.PullRequest.GetNumber()
	}

	if prNum != 0 {
		if data.PullRequest.PullRequest == nil {
			pr, _, err := handler.GitHub.PullRequests.Get(ctx, data.Repository.Owner, data.Repository.Name, prNum)
			if err != nil {
				return fmt.Errorf("get a pull request: %w", err)
			}
			handler.setPullRequest(&data, pr)
		}

		files, _, err := getPRFiles(ctx, handler.GitHub, data.Repository.Owner, data.Repository.Name, prNum, data.PullRequest.PullRequest.GetChangedFiles())
		if err != nil {
			return fmt.Errorf("get pull request files: %w", err)
		}
		data.PullRequest.ChangedFileNames = extractPRFileNames(files)

		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String("LAMBUILD_PR_NUMBER"),
			Value: aws.String(strconv.Itoa(prNum)),
		}, &codebuild.EnvironmentVariable{
			Name:  aws.String("LAMBUILD_PR_AUTHOR"),
			Value: aws.String(data.PullRequest.PullRequest.GetUser().GetLogin()),
		}, &codebuild.EnvironmentVariable{
			Name:  aws.String("LAMBUILD_PR_BASE_REF"),
			Value: aws.String(data.PullRequest.PullRequest.GetBase().GetRef()),
		}, &codebuild.EnvironmentVariable{
			Name:  aws.String("LAMBUILD_PR_HEAD_REF"),
			Value: aws.String(data.PullRequest.PullRequest.GetHead().GetRef()),
		})
	}

	if len(buildspec.Batch.BuildGraph) != 0 {
		return handler.handleGraph(ctx, logE, data, buildspec, repo, envs)
	}
	if len(buildspec.Batch.BuildList) != 0 {
		return handler.handleList(ctx, logE, data, buildspec, repo, envs)
	}
	if !buildspec.Batch.BuildMatrix.Empty() {
		return handler.handleMatrix(ctx, logE, data, buildspec, repo, envs)
	}

	return nil
}

func (handler *Handler) setPullRequest(data *Data, pr *github.PullRequest) {
	data.Ref = pr.GetHead().GetRef()
	data.PullRequest.LabelNames = extractLabelNames(pr.Labels)
	data.PullRequest.PullRequest = pr
}

func (handler *Handler) handlePushEvent(ctx context.Context, event Event, pushEvent *github.PushEvent) error {
	repo := pushEvent.GetRepo()
	data := Data{
		Event: event,
		Repository: Repository{
			FullName: repo.GetFullName(),
			Name:     repo.GetName(),
		},
		HeadCommitMessage: pushEvent.GetHeadCommit().GetMessage(),
		SHA:               pushEvent.GetAfter(),
		Ref:               pushEvent.GetRef(),
	}
	return handler.handleEvent(ctx, data)
}

func (handler *Handler) handlePREvent(ctx context.Context, event Event, prEvent *github.PullRequestEvent) error {
	repo := prEvent.GetRepo()
	pr := prEvent.GetPullRequest()

	data := Data{
		Event: event,
		Repository: Repository{
			FullName: repo.GetFullName(),
			Name:     repo.GetName(),
		},
		SHA: prEvent.GetAfter(),
	}
	handler.setPullRequest(&data, pr)
	return handler.handleEvent(ctx, data)
}

func (handler *Handler) Do(ctx context.Context, event Event) error {
	if err := github.ValidateSignature(event.Headers.Signature, []byte(event.Body), []byte(handler.Secret.WebhookSecret)); err != nil {
		// TODO return 400
		fmt.Println(err)
		return nil
	}
	body, err := github.ParseWebHook(event.Headers.Event, []byte(event.Body))
	if err != nil {
		return fmt.Errorf("parse a webhook payload: %w", err)
	}

	switch event.Headers.Event {
	case "push":
		return handler.handlePushEvent(ctx, event, body.(*github.PushEvent))
	case "pull_request":
		return handler.handlePREvent(ctx, event, body.(*github.PullRequestEvent))
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
	if secretNameWebhookSecret == "" {
		return errors.New("the environment variable 'SSM_PARAMETER_NAME_WEBHOOK_SECRET' is required")
	}

	if context := os.Getenv("BUILD_STATUS_CONTEXT"); context != "" {
		tpl, err := template.New("_").Funcs(sprig.TxtFuncMap()).Parse(context)
		if err != nil {
			return fmt.Errorf("parse BUILD_STATUS_CONTEXT as template (%s): %w", context, err)
		}
		handler.BuildStatusContext = tpl
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
	return nil
}
