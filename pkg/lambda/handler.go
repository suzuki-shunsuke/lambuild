package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
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

func (handler *Handler) getRepo(body wh.Body) (config.Repository, bool) {
	for _, repo := range handler.Config.Repositories {
		if repo.Name == body.Repository.FullName {
			return repo, true
		}
	}
	return config.Repository{}, false
}

func (handler *Handler) getHook(webhook wh.Webhook, body wh.Body, repo config.Repository) (config.Hook, bool, error) {
	for _, hook := range repo.Hooks {
		if hook.Event != webhook.Headers.Event {
			continue
		}
		f, err := handler.MatchfileParser.Match([]string{body.Ref}, hook.RefConditions)
		if err != nil {
			return hook, false, err
		}
		if f {
			return hook, true, nil
		}
	}
	return config.Hook{}, false, nil
}

func (handler *Handler) Do(ctx context.Context, webhook wh.Webhook) error {
	if err := github.ValidateSignature(webhook.Headers.Signature, []byte(webhook.Body), []byte(handler.Secret.WebhookSecret)); err != nil {
		// TODO return 400
		fmt.Println(err)
		return nil
	}
	body := wh.Body{}
	if err := json.Unmarshal([]byte(webhook.Body), &body); err != nil {
		return err
	}

	logE := logrus.WithFields(logrus.Fields{
		"repo": body.Repository.FullName,
		"ref":  body.Ref,
	})

	repo, f := handler.getRepo(body)
	if !f {
		logE.Debug("no repo matches")
		return nil
	}

	hook, f, err := handler.getHook(webhook, body, repo)
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

	owner := strings.Split(body.Repository.FullName, "/")[0]
	file, _, _, err := handler.GitHub.Repositories.GetContents(ctx, owner, body.Repository.Name, hook.Config, &github.RepositoryContentGetOptions{Ref: body.Ref})
	if err != nil {
		// start build
		logE.WithError(err).Debug("no content is found")
		return nil
	}
	content, err := file.GetContent()
	if err != nil {
		return fmt.Errorf("get a content: %w", err)
	}

	// TODO generate buildspec

	buildOut, err := handler.CodeBuild.StartBuildBatchWithContext(ctx, &codebuild.StartBuildBatchInput{
		BuildspecOverride: aws.String(content),
		ProjectName:       aws.String(repo.CodeBuild.ProjectName),
		SourceVersion:     aws.String(body.After),
	})
	if err != nil {
		return fmt.Errorf("start a batch build: %w", err)
	}
	logE.WithFields(logrus.Fields{
		"build_arn": *buildOut.BuildBatch.Arn,
	}).Info("start a build")
	return nil
}

func (handler *Handler) Init(ctx context.Context) error {
	cfg := config.Config{}
	configRaw := os.Getenv("CONFIG")
	if err := yaml.Unmarshal([]byte(configRaw), &cfg); err != nil {
		return err
	}
	handler.Config = cfg
	handler.Region = os.Getenv("REGION")
	sess := session.Must(session.NewSession())
	if secretNameGitHubToken := os.Getenv("SSM_PARAMETER_NAME_GITHUB_TOKEN"); secretNameGitHubToken != "" {
		if secretNameWebhookSecret := os.Getenv("SSM_PARAMETER_NAME_WEBHOOK_SECRET"); secretNameWebhookSecret != "" {
			if err := handler.readSecretFromSSM(ctx, sess, secretNameGitHubToken, secretNameWebhookSecret); err != nil {
				return err
			}
		}
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
			rawConditions := strings.Split(strings.TrimSpace(hook.Refs), "\n")
			conditions, err := handler.MatchfileParser.ParseConditions(rawConditions)
			if err != nil {
				return err
			}
			hook.RefConditions = conditions
			repo.Hooks[j] = hook
		}
		cfg.Repositories[i] = repo
	}

	return nil
}
