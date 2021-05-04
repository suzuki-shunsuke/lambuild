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

const defaultErrorNotificationTemplate = `
lambuild failed to procceed the request.
Please check.

` + "```" + `
{{.Error}}
` + "```"

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

// Do is the Lambda Function's endpoint.
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
		Event:  event,
		GitHub: handler.GitHub,
	}

	if event.Headers.Event != "push" && event.Headers.Event != "pull_request" {
		// Events other than "push" and "pull_request" aren't supported.
		// These events are ignored.
		return nil
	}

	switch event.Headers.Event {
	case "push":
		pushEvent := body.(*github.PushEvent) //nolint:forcetypeassert
		repo := pushEvent.GetRepo()
		data.Repository = Repository{
			FullName: repo.GetFullName(),
			Name:     repo.GetName(),
		}
		data.HeadCommitMessage = pushEvent.GetHeadCommit().GetMessage()
		data.SHA = pushEvent.GetAfter()
		data.Ref = pushEvent.GetRef()
	case "pull_request":
		prEvent := body.(*github.PullRequestEvent) //nolint:forcetypeassert

		repo := prEvent.GetRepo()
		pr := prEvent.GetPullRequest()
		data.Repository = Repository{
			FullName: repo.GetFullName(),
			Name:     repo.GetName(),
		}
		data.SHA = prEvent.GetAfter()
		data.Ref = pr.GetHead().GetRef()
		data.PullRequest.LabelNames = extractLabelNames(pr.Labels)
		data.PullRequest.PullRequest = pr
	}
	data.Repository.Owner = strings.Split(data.Repository.FullName, "/")[0]
	if err := handler.handleEvent(ctx, data); err != nil {
		handler.sendErrorNotificaiton(ctx, err, data.Repository.Owner, data.Repository.Name, data.GetPRNumber(), data.SHA)
		return err
	}
	return nil
}

// getRepo returns the configuration of given repository name.
// If no configuration is found, the second returned value is false.
func (handler *Handler) getRepo(repoName string) (config.Repository, bool) {
	for _, repo := range handler.Config.Repositories {
		if repo.Name == repoName {
			return repo, true
		}
	}
	return config.Repository{}, false
}

// matchHook returns true if data matches hook's condition.
func (handler *Handler) matchHook(data *Data, hook config.Hook) (bool, error) {
	if hook.If == nil {
		return true, nil
	}
	f, err := runExpr(hook.If, data)
	if err != nil {
		return false, fmt.Errorf("evaluate an expression: %w", err)
	}
	return f.(bool), nil
}

// getHook returns a hook configuration which data matches.
// If data doesn't match any configuration, the second returned value is false.
func (handler *Handler) getHook(data *Data, repo config.Repository) (config.Hook, bool, error) {
	for _, hook := range repo.Hooks {
		f, err := handler.matchHook(data, hook)
		if err != nil {
			return config.Hook{}, false, err
		}
		if f {
			return hook, true, nil
		}
	}
	return config.Hook{}, false, nil
}

// getConfigFromRepo gets the configuration file from the target repository
func (handler *Handler) getConfigFromRepo(ctx context.Context, logE *logrus.Entry, data *Data, hook config.Hook) (bspec.Buildspec, error) {
	buildspec := bspec.Buildspec{}
	// get the configuration file from the target repository
	if hook.Config == "" {
		// set the default value
		hook.Config = "lambuild.yaml"
	}
	file, _, _, err := handler.GitHub.Repositories.GetContents(ctx, data.Repository.Owner, data.Repository.Name, hook.Config, &github.RepositoryContentGetOptions{Ref: data.Ref})
	if err != nil {
		logE.WithFields(logrus.Fields{
			"path": hook.Config,
		}).WithError(err).Error("get the configuration file")
		return buildspec, nil
	}
	if file == nil {
		logE.Warn("downloaded file is nil")
		return buildspec, nil
	}
	content, err := file.GetContent()
	if err != nil {
		return buildspec, fmt.Errorf("get a content: %w", err)
	}

	if err := yaml.Unmarshal([]byte(content), &buildspec); err != nil {
		return buildspec, fmt.Errorf("unmarshal a buildspec: %w", err)
	}
	data.Lambuild = buildspec.Lambuild
	buildspec.Lambuild = bspec.Lambuild{}
	return buildspec, nil
}

type BuildInput struct {
	Build      *codebuild.StartBuildInput
	BatchBuild *codebuild.StartBuildBatchInput
	Batched    bool
	Empty      bool
}

func (handler *Handler) handleEvent(ctx context.Context, data *Data) error {
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

	// get the configuration file from the target repository
	buildspec, err := handler.getConfigFromRepo(ctx, logE, data, hook)
	if err != nil {
		return err
	}

	buildInput := &BuildInput{
		Build: &codebuild.StartBuildInput{
			ProjectName:   aws.String(repo.CodeBuild.ProjectName),
			SourceVersion: aws.String(data.SHA),
		},
		BatchBuild: &codebuild.StartBuildBatchInput{
			ProjectName:   aws.String(repo.CodeBuild.ProjectName),
			SourceVersion: aws.String(data.SHA),
		},
	}

	if len(buildspec.Batch.BuildGraph) != 0 {
		if err := handler.handleGraph(buildInput, logE, data, buildspec); err != nil {
			return err
		}
	}

	if len(buildspec.Batch.BuildList) != 0 {
		if err := handler.handleList(buildInput, logE, data, buildspec); err != nil {
			return err
		}
	}

	if !buildspec.Batch.BuildMatrix.Empty() {
		if err := handler.handleMatrix(buildInput, logE, data, buildspec); err != nil {
			return err
		}
	}

	if buildInput.Empty {
		return nil
	}

	if buildInput.Batched {
		buildOut, err := handler.CodeBuild.StartBuildBatchWithContext(ctx, buildInput.BatchBuild)
		if err != nil {
			return fmt.Errorf("start a batch build: %w", err)
		}
		logE.WithFields(logrus.Fields{
			"build_arn": *buildOut.BuildBatch.Arn,
		}).Info("start a batch build")
		return nil
	}

	buildOut, err := handler.CodeBuild.StartBuildWithContext(ctx, buildInput.Build)
	if err != nil {
		return fmt.Errorf("start a batch build: %w", err)
	}
	logE.WithFields(logrus.Fields{
		"build_arn": *buildOut.Build.Arn,
	}).Info("start a build")

	return nil
}

// sendErrorNotificaiton sends a comment to GitHub PullRequest or commit to notify an error.
// If prNumber isn't zero a comment is sent to the pull reqquest.
// If prNumber is zero, which means the event isn't associated with any pull request, a comment is sent to a comment.
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
	var cmt string
	if renderErr := handler.ErrorNotificationTemplate.Execute(buf, map[string]interface{}{
		"Error": e,
	}); renderErr != nil {
		logE.WithError(renderErr).Error("render a comment to send it to the pull request")
		cmt = "lambuild failed to procceed the request: " + e.Error()
	} else {
		cmt = buf.String()
	}

	if prNumber == 0 {
		// send a comment to commit
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

	if cntxt := os.Getenv("BUILD_STATUS_CONTEXT"); cntxt != "" {
		tpl, err := template.New("_").Funcs(sprig.TxtFuncMap()).Parse(cntxt)
		if err != nil {
			return fmt.Errorf("parse BUILD_STATUS_CONTEXT as template (%s): %w", cntxt, err)
		}
		handler.BuildStatusContext = tpl
	}

	if err := handler.readSecretFromSSM(ctx, sess, secretNameGitHubToken, secretNameWebhookSecret); err != nil {
		return err
	}

	errTpl := os.Getenv("ERROR_NOTIFICATION_TEMPLATE")
	if errTpl == "" {
		errTpl = defaultErrorNotificationTemplate
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
