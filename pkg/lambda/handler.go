package lambda

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
	generator "github.com/suzuki-shunsuke/lambuild/pkg/build-input-generator"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"gopkg.in/yaml.v2"
)

type Handler struct {
	Config    config.Config
	Secret    Secret
	GitHub    *github.Client
	CodeBuild *codebuild.CodeBuild
}

type Secret struct {
	GitHubToken   string `json:"github_token"`
	WebhookSecret string `json:"webhook_secret"`
}

// Do is the Lambda Function's endpoint.
func (handler *Handler) Do(ctx context.Context, event domain.Event) error {
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

	data := &domain.Data{
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
		data.Repository = domain.Repository{
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
		data.Repository = domain.Repository{
			FullName: repo.GetFullName(),
			Name:     repo.GetName(),
		}
		data.SHA = prEvent.GetAfter()
		data.Ref = pr.GetHead().GetRef()
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
func matchHook(data *domain.Data, hook config.Hook) (bool, error) {
	if hook.If == nil {
		return true, nil
	}
	f, err := domain.RunExpr(hook.If, data)
	if err != nil {
		return false, fmt.Errorf("evaluate an expression: %w", err)
	}
	return f.(bool), nil
}

// getHook returns a hook configuration which data matches.
// If data doesn't match any configuration, the second returned value is false.
func getHook(data *domain.Data, repo config.Repository) (config.Hook, bool, error) {
	for _, hook := range repo.Hooks {
		f, err := matchHook(data, hook)
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
func (handler *Handler) getConfigFromRepo(ctx context.Context, logE *logrus.Entry, data *domain.Data, hook config.Hook) (bspec.Buildspec, error) {
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
		}).WithError(err).Error("")
		return buildspec, fmt.Errorf("get a configuration file by GitHub API: %w", err)
	}
	content, err := file.GetContent()
	if err != nil {
		return buildspec, fmt.Errorf("get a content: %w", err)
	}

	m := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(content), &m); err != nil {
		return buildspec, fmt.Errorf("unmarshal a buildspec to map: %w", err)
	}

	if err := yaml.Unmarshal([]byte(content), &buildspec); err != nil {
		return buildspec, fmt.Errorf("unmarshal a buildspec: %w", err)
	}
	data.Lambuild = buildspec.Lambuild
	buildspec.Lambuild = bspec.Lambuild{}
	buildspec.Map = m
	return buildspec, nil
}

func (handler *Handler) handleEvent(ctx context.Context, data *domain.Data) error {
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

	hook, f, err := getHook(data, repo)
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
	logE.Debug("get a configuration file from the source repository")

	buildInput, err := generator.GenerateInput(logE, handler.Config.BuildStatusContext, data, buildspec, repo)
	if err != nil {
		return fmt.Errorf("generate a build input: %w", err)
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
	if renderErr := handler.Config.ErrorNotificationTemplate.Execute(buf, map[string]interface{}{
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
