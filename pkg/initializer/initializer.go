package initializer

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	gh "github.com/suzuki-shunsuke/lambuild/pkg/github"
	"github.com/suzuki-shunsuke/lambuild/pkg/lambda"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
	"gopkg.in/yaml.v2"
)

func InitializeHandler(ctx context.Context, handler *lambda.Handler) error {
	cfg := config.Config{}
	if err := readConfigFromSource(ctx, &cfg); err != nil {
		return fmt.Errorf("read configuration from source: %w", err)
	}

	if lvl := cfg.LogLevel.Get(); lvl != 0 {
		logrus.SetLevel(lvl)
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
	}

	if cfg.SSMParameter.ParameterName.WebhookSecret == "" {
		cfg.SSMParameter.ParameterName.WebhookSecret = os.Getenv("SSM_PARAMETER_NAME_WEBHOOK_SECRET")
	}

	if cfg.SecretsManager.SecretID == "" {
		cfg.SecretsManager.SecretID = os.Getenv("SECRETS_MANAGER_SECRET_ID")
	}

	if cfg.SecretsManager.VersionID == "" {
		cfg.SecretsManager.VersionID = os.Getenv("SECRETS_MANAGER_VERSION_ID")
	}

	if cfg.BuildStatusContext.Empty() {
		if cntxt := os.Getenv("BUILD_STATUS_CONTEXT"); cntxt != "" {
			tpl, err := template.New(cntxt)
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
	switch {
	case cfg.SSMParameter.ParameterName.GitHubToken != "":
		ssmSvc := ssm.New(sess, aws.NewConfig().WithRegion(handler.Config.Region))
		secret, err := readSecretFromSSM(ctx, ssmSvc, handler.Config.SSMParameter.ParameterName)
		if err != nil {
			return err
		}
		handler.Secret = secret
	case cfg.SecretsManager.SecretID != "":
		svc := secretsmanager.New(sess, aws.NewConfig().WithRegion(handler.Config.Region))
		secret, err := readSecretFromSecretsManager(ctx, svc, handler.Config.SecretsManager)
		if err != nil {
			return err
		}
		handler.Secret = secret
	default:
		return errors.New("secrets aren't configured")
	}

	ghClient := gh.New(ctx, handler.Secret.GitHubToken)
	handler.GitHub = &ghClient
	handler.CodeBuild = codebuild.New(sess, aws.NewConfig().WithRegion(handler.Config.Region))

	// get AWS Account ID
	stsSvc := sts.New(sess, aws.NewConfig().WithRegion(handler.Config.Region))
	stsOutput, err := stsSvc.GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("get a caller identity: %w", err)
	}
	handler.AWSAccountID = aws.StringValue(stsOutput.Account)

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
			for _, hook := range repo.Hooks {
				if hook.ProjectName == "" {
					return fmt.Errorf(`'project-name' is required (repo: %s)`, repo.Name)
				}
			}
		}
	}
	return nil
}

func readConfigFromSource(ctx context.Context, cfg *config.Config) error {
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
		if err := readAppConfig(ctx, cfg); err != nil {
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
	if !cfg.ErrorNotificationTemplate.Empty() {
		return nil
	}
	if errTpl := os.Getenv("ERROR_NOTIFICATION_TEMPLATE"); errTpl != "" {
		tpl, err := template.New(errTpl)
		if err != nil {
			return fmt.Errorf("parse ERROR_NOTIFICATION_TEMPLATE as template: %w", err)
		}
		cfg.ErrorNotificationTemplate = tpl
		return nil
	}
	tpl, err := template.New(defaultErrorNotificationTemplate)
	if err != nil {
		return fmt.Errorf("parse defaultErroNotificationTemplate as template: %w", err)
	}
	cfg.ErrorNotificationTemplate = tpl
	return nil
}
