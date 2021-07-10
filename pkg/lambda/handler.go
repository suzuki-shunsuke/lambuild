package lambda

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-github/v37/github"
	"github.com/sirupsen/logrus"
	generator "github.com/suzuki-shunsuke/lambuild/pkg/build-input-generator"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"golang.org/x/sync/errgroup"
)

type Handler struct {
	Config       config.Config
	Secret       Secret
	GitHub       domain.GitHub
	CodeBuild    CodeBuild
	AWSAccountID string
}

type CodeBuild interface {
	StartBuildBatchWithContext(ctx aws.Context, input *codebuild.StartBuildBatchInput, opts ...request.Option) (*codebuild.StartBuildBatchOutput, error)
	StartBuildWithContext(ctx aws.Context, input *codebuild.StartBuildInput, opts ...request.Option) (*codebuild.StartBuildOutput, error)
}

type Secret struct {
	GitHubToken   string `json:"github-token"`
	WebhookSecret string `json:"webhook-secret"`
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

	data := domain.NewData()
	data.Event = event
	data.GitHub = handler.GitHub
	data.AWS.Region = handler.Config.Region
	data.AWS.AccountID = handler.AWSAccountID

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
		data.HeadCommitMessage.Set(pushEvent.GetHeadCommit().GetMessage())
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
		data.PullRequest.PullRequest.Set(pr)
	}
	data.Repository.Owner = strings.Split(data.Repository.FullName, "/")[0]
	if err := handler.handleEvent(ctx, &data); err != nil {
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

	data.AWS.CodeBuildProjectName = repo.CodeBuild.ProjectName

	hook, f, err := getHook(data, repo)
	if err != nil {
		return err
	}
	if !f {
		logE.Debug("no hook matches")
		return nil
	}
	if hook.ProjectName != "" {
		data.AWS.CodeBuildProjectName = hook.ProjectName
	}
	logE = logE.WithFields(logrus.Fields{
		"config": hook.Config,
	})

	// get the configuration files from the target repository
	buildspecs, err := handler.getConfigFromRepo(ctx, logE, data, hook)
	if err != nil {
		return err
	}
	logE.WithFields(logrus.Fields{
		"number_of_buildspecs": len(buildspecs),
	}).Debug("get configuration files from the source repository")

	var eg errgroup.Group
	for _, buildspec := range buildspecs {
		buildspec := buildspec
		eg.Go(func() error {
			return handler.handleBuildspec(ctx, logE, data, buildspec, repo, hook)
		})
	}
	return eg.Wait() //nolint:wrapcheck
}

func (handler *Handler) handleBuildspec(ctx context.Context, logE *logrus.Entry, data *domain.Data, buildspec bspec.Buildspec, repo config.Repository, hook config.Hook) error {
	buildInput, err := generator.GenerateInput(logE, handler.Config.BuildStatusContext, data, buildspec, repo)
	if err != nil {
		logE.WithError(err).Error("generate a build input")
		return fmt.Errorf("generate a build input: %w", err)
	}

	if buildInput.Empty {
		return nil
	}

	projectName := repo.CodeBuild.ProjectName
	if hook.ProjectName != "" {
		projectName = hook.ProjectName
	}

	cb := handler.CodeBuild
	assumeRoleARN := repo.CodeBuild.AssumeRoleARN
	if hook.AssumeRoleARN != "" {
		assumeRoleARN = hook.AssumeRoleARN
	}
	if assumeRoleARN != "" {
		sess := session.Must(session.NewSession())
		creds := stscreds.NewCredentials(sess, assumeRoleARN)
		cb = codebuild.New(sess, &aws.Config{Credentials: creds, Region: aws.String(handler.Config.Region)})
	}

	if buildInput.Batched {
		buildInput.BatchBuild.ProjectName = aws.String(projectName)
		buildInput.BatchBuild.SourceVersion = aws.String(data.SHA)
		if hook.ServiceRole != "" {
			buildInput.BatchBuild.ServiceRoleOverride = aws.String(hook.ServiceRole)
		}
		buildOut, err := cb.StartBuildBatchWithContext(ctx, buildInput.BatchBuild)
		if err != nil {
			logE.WithError(err).Error("start a batch build")
			return fmt.Errorf("start a batch build: %w", err)
		}
		logE.WithFields(logrus.Fields{
			"build_arn": *buildOut.BuildBatch.Arn,
		}).Info("start a batch build")
		return nil
	}

	for _, build := range buildInput.Builds {
		build.ProjectName = aws.String(projectName)
		build.SourceVersion = aws.String(data.SHA)
		if hook.ServiceRole != "" {
			build.ServiceRoleOverride = aws.String(hook.ServiceRole)
		}
		buildOut, err := cb.StartBuildWithContext(ctx, build)
		if err != nil {
			logE.WithError(err).Error("start a build")
			return fmt.Errorf("start a build: %w", err)
		}
		logE.WithFields(logrus.Fields{
			"build_arn": *buildOut.Build.Arn,
		}).Info("start a build")
	}
	return nil
}
