package lambda

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
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
	Config                    config.Config
	Secret                    Secret
	SecretID                  string
	SecretVersionID           string
	Region                    string
	BuildStatusContext        *template.Template
	GitHub                    *github.Client
	CodeBuild                 *codebuild.CodeBuild
	ErrorNotificationTemplate *template.Template
}

func (handler *Handler) Do(ctx context.Context, event Event) error {
	if err := github.ValidateSignature(event.Headers.Signature, []byte(event.Body), []byte(handler.Secret.WebhookSecret)); err != nil {
		// TODO return 400
		logrus.Debug(err)
		return nil
	}
	body, err := github.ParseWebHook(event.Headers.Event, []byte(event.Body))
	if err != nil {
		return fmt.Errorf("parse a webhook payload: %w", err)
	}
	event.Payload = body

	data := &Data{
		Event: event,
	}

	if event.Headers.Event != "push" && event.Headers.Event != "pull_request" {
		// Events other than "push" and "pull_request" aren't supported.
		// These events are ignored.
		return nil
	}

	switch event.Headers.Event {
	case "push":
		pushEvent := body.(*github.PushEvent)
		repo := pushEvent.GetRepo()
		data.Repository = Repository{
			FullName: repo.GetFullName(),
			Name:     repo.GetName(),
		}
		data.HeadCommitMessage = pushEvent.GetHeadCommit().GetMessage()
		data.SHA = pushEvent.GetAfter()
		data.Ref = pushEvent.GetRef()
	case "pull_request":
		prEvent := body.(*github.PullRequestEvent)

		repo := prEvent.GetRepo()
		pr := prEvent.GetPullRequest()
		data.Repository = Repository{
			FullName: repo.GetFullName(),
			Name:     repo.GetName(),
		}
		data.SHA = prEvent.GetAfter()
		handler.setPullRequest(data, pr)
	}
	if err := handler.handleEvent(ctx, data); err != nil {
		handler.sendErrorNotificaiton(ctx, err, data.Repository.Owner, data.Repository.Name, data.GetPRNumber(), data.SHA)
		return err
	}
	return nil
}

func (handler *Handler) getRepo(repoName string) (config.Repository, bool) {
	for _, repo := range handler.Config.Repositories {
		if repo.Name == repoName {
			return repo, true
		}
	}
	return config.Repository{}, false
}

func (handler *Handler) matchHook(data *Data, repo config.Repository, hook config.Hook) (bool, error) {
	if hook.If == nil {
		return true, nil
	}
	f, err := runExpr(hook.If, data)
	if err != nil {
		return false, fmt.Errorf("evaluate an expression: %w", err)
	}
	return f.(bool), nil
}

func (handler *Handler) getHook(data *Data, repo config.Repository) (config.Hook, bool, error) {
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

func (handler *Handler) handleEvent(ctx context.Context, data *Data) error {
	data.Repository.Owner = strings.Split(data.Repository.FullName, "/")[0]
	data.GitHub = handler.GitHub
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
	data.Lambuild = buildspec.Lambuild
	buildspec.Lambuild = bspec.Lambuild{}

	if len(buildspec.Batch.BuildGraph) != 0 {
		return handler.handleGraph(ctx, logE, data, buildspec, repo)
	}
	if len(buildspec.Batch.BuildList) != 0 {
		return handler.handleList(ctx, logE, data, buildspec, repo)
	}
	if !buildspec.Batch.BuildMatrix.Empty() {
		return handler.handleMatrix(ctx, logE, data, buildspec, repo)
	}

	return nil
}

func (handler *Handler) setPullRequest(data *Data, pr *github.PullRequest) {
	data.Ref = pr.GetHead().GetRef()
	data.PullRequest.LabelNames = extractLabelNames(pr.Labels)
	data.PullRequest.PullRequest = pr
}

const errorNotificationTemplate = `
lambuild failed to procceed the request.
Please check.

` + "```" + `
{{.Error}}
` + "```"

func (handler *Handler) sendErrorNotificaiton(ctx context.Context, e error, repoOwner, repoName string, prNumber int, sha string) {
	logE := logrus.WithFields(logrus.Fields{
		"original_error": e,
		"repo_owner":     repoOwner,
		"repo_name":      repoName,
		"pr_number":      prNumber,
		"sha":            sha,
	})
	// generate a comment
	buf := &bytes.Buffer{}
	cmt := ""
	if renderErr := handler.ErrorNotificationTemplate.Execute(buf, map[string]interface{}{
		"Error": e,
	}); renderErr != nil {
		logE.WithError(renderErr).Error("render a comment to send it to the pull request")
		cmt = "lambuild failed to procceed the request: " + e.Error()
	} else {
		cmt = buf.String()
	}
	if prNumber == 0 {
		if _, _, cmtErr := handler.GitHub.Repositories.CreateComment(ctx, repoOwner, repoName, sha, &github.RepositoryComment{
			Body: github.String(cmt),
		}); cmtErr != nil {
			logE.WithError(cmtErr).Error("send a comment to the commit")
		}
		logE.Info("send a comment to the commit")
		return
	}
	// send a comment to pull request
	if _, _, cmtErr := handler.GitHub.Issues.CreateComment(ctx, repoOwner, repoName, prNumber, &github.IssueComment{
		Body: github.String(cmt),
	}); cmtErr != nil {
		logE.WithError(cmtErr).Error("send a comment to the pull request")
	}
	logE.Info("send a comment to the pull request")
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

	errTpl := os.Getenv("ERROR_NOTIFICATION_TEMPLATE")
	if errTpl == "" {
		errTpl = errorNotificationTemplate
	}

	tpl, err := template.New("_").Parse(errTpl)
	if err != nil {
		return fmt.Errorf("parse an error notification template: %w", err)
	}
	handler.ErrorNotificationTemplate = tpl

	handler.GitHub = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: handler.Secret.GitHubToken},
	)))
	handler.CodeBuild = codebuild.New(sess, aws.NewConfig().WithRegion(handler.Region))
	return nil
}
