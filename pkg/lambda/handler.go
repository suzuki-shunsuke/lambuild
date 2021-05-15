package lambda

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
	generator "github.com/suzuki-shunsuke/lambuild/pkg/build-input-generator"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"golang.org/x/sync/errgroup"
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

func (handler *Handler) handleEvent(ctx context.Context, data *domain.Data) error {
	logE := logrus.WithFields(logrus.Fields{
		"repo_full_name": data.Repository.FullName,
		"repo_owner":     data.Repository.Owner,
		"repo_name":      data.Repository.Name,
		"ref":            data.Ref,
	})

	repo, f := getRepo(handler.Config.Repositories, data.Repository.FullName)
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

	// get the configuration files from the target repository
	buildspecs, err := handler.getConfigFromRepo(ctx, logE, data, hook)
	if err != nil {
		return err
	}
	logE.Debug("get a configuration file from the source repository")

	var eg errgroup.Group
	for _, buildspec := range buildspecs {
		buildspec := buildspec
		eg.Go(func() error {
			buildInput, err := generator.GenerateInput(logE, handler.Config.BuildStatusContext, data, buildspec, repo)
			if err != nil {
				logE.WithError(err).Error("generate a build input")
				return fmt.Errorf("generate a build input: %w", err)
			}

			if buildInput.Empty {
				return nil
			}

			if buildInput.Batched {
				buildOut, err := handler.CodeBuild.StartBuildBatchWithContext(ctx, buildInput.BatchBuild)
				if err != nil {
					logE.WithError(err).Error("start a batch build")
					return fmt.Errorf("start a batch build: %w", err)
				}
				logE.WithFields(logrus.Fields{
					"build_arn": *buildOut.BuildBatch.Arn,
				}).Info("start a batch build")
				return nil
			}

			buildOut, err := handler.CodeBuild.StartBuildWithContext(ctx, buildInput.Build)
			if err != nil {
				logE.WithError(err).Error("start a build")
				return fmt.Errorf("start a build: %w", err)
			}
			logE.WithFields(logrus.Fields{
				"build_arn": *buildOut.Build.Arn,
			}).Info("start a build")
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
