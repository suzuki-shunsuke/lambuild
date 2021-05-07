package lambda

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	templ "github.com/suzuki-shunsuke/lambuild/pkg/template"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

func (handler *Handler) Init(ctx context.Context) error {
	cfg := config.Config{}
	if err := handler.readConfigFromSource(ctx, &cfg); err != nil {
		return fmt.Errorf("read configuration from source: %w", err)
	}

	if cfg.LogLevel != 0 {
		logrus.SetLevel(cfg.LogLevel)
	}

	if err := validateRepositories(cfg.Repositories); err != nil {
		return fmt.Errorf("validate repositories: %w", err)
	}

	if cfg.Region == "" {
		cfg.Region = os.Getenv("REGION")
		if cfg.Region == "" {
			return errors.New("the configuration 'region' is required")
		}
	}

	if cfg.SSMParameter.ParameterName.GitHubToken == "" {
		cfg.SSMParameter.ParameterName.GitHubToken = os.Getenv("SSM_PARAMETER_NAME_GITHUB_TOKEN")
		if cfg.SSMParameter.ParameterName.GitHubToken == "" {
			return errors.New("the configuration 'SSM_PARAMETER_NAME_GITHUB_TOKEN' is required")
		}
	}

	if cfg.SSMParameter.ParameterName.WebhookSecret == "" {
		cfg.SSMParameter.ParameterName.WebhookSecret = os.Getenv("SSM_PARAMETER_NAME_WEBHOOK_SECRET")
		if cfg.SSMParameter.ParameterName.WebhookSecret == "" {
			return errors.New("the configuration 'SSM_PARAMETER_NAME_WEBHOOK_SECRET' is required")
		}
	}

	if cfg.BuildStatusContext == nil {
		if cntxt := os.Getenv("BUILD_STATUS_CONTEXT"); cntxt != "" {
			tpl, err := templ.Compile(cntxt)
			if err != nil {
				return fmt.Errorf("parse BUILD_STATUS_CONTEXT as template (%s): %w", cntxt, err)
			}
			cfg.BuildStatusContext = tpl
		}
	}

	if err := setErrorNotificationTemplate(&cfg); err != nil {
		return fmt.Errorf("configure error notification template: %w", err)
	}

	handler.Config = cfg

	sess := session.Must(session.NewSession())
	if err := handler.readSecretFromSSM(ctx, sess); err != nil {
		return err
	}

	handler.GitHub = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: handler.Secret.GitHubToken},
	)))
	handler.CodeBuild = codebuild.New(sess, aws.NewConfig().WithRegion(handler.Config.Region))
	return nil
}

func validateRepositories(repos []config.Repository) error {
	if len(repos) == 0 {
		return errors.New(`the configuration 'repositories' is required`)
	}
	for _, repo := range repos {
		if repo.Name == "" {
			return errors.New(`the repository 'name' is required`)
		}
		if repo.CodeBuild.ProjectName == "" {
			return fmt.Errorf(`'project-name' is required (repo: %s)`, repo.Name)
		}
	}
	return nil
}

func (handler *Handler) readConfigFromSource(ctx context.Context, cfg *config.Config) error {
	switch cfgSrc := os.Getenv("CONFIG_SOURCE"); cfgSrc {
	case "", "env":
		configRaw := os.Getenv("CONFIG")
		if configRaw == "" {
			return errors.New("the environment variable 'CONFIG' is required")
		}
		if err := yaml.Unmarshal([]byte(configRaw), cfg); err != nil {
			return fmt.Errorf("parse the environment variable 'CONFIG' as YAML: %w", err)
		}
	case "appconfig-extension":
		if err := handler.readAppConfig(ctx, cfg); err != nil {
			return fmt.Errorf("read application configuration from AppConfig: %w", err)
		}
	default:
		return errors.New("CONFIG_SOURCE is invalid: " + cfgSrc)
	}
	return nil
}

const defaultErrorNotificationTemplate = `
lambuild failed to procceed the request.
Please check.

` + "```" + `
{{.Error}}
` + "```"

func setErrorNotificationTemplate(cfg *config.Config) error {
	if cfg.ErrorNotificationTemplate != nil {
		return nil
	}
	if errTpl := os.Getenv("ERROR_NOTIFICATION_TEMPLATE"); errTpl != "" {
		tpl, err := templ.Compile(errTpl)
		if err != nil {
			return fmt.Errorf("parse ERROR_NOTIFICATION_TEMPLATE as template: %w", err)
		}
		cfg.ErrorNotificationTemplate = tpl
		return nil
	}
	tpl, err := templ.Compile(defaultErrorNotificationTemplate)
	if err != nil {
		return fmt.Errorf("parse defaultErroNotificationTemplate as template: %w", err)
	}
	cfg.ErrorNotificationTemplate = tpl
	return nil
}
