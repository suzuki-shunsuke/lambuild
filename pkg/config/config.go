package config

import (
	"fmt"
	"text/template"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/sirupsen/logrus"
	templ "github.com/suzuki-shunsuke/lambuild/pkg/template"
)

type Config struct {
	Region                    string
	Repositories              []Repository
	LogLevel                  logrus.Level       `yaml:"-"`
	BuildStatusContext        *template.Template `yaml:"-"`
	ErrorNotificationTemplate *template.Template `yaml:"-"`
	SSMParameter              SSMParameter       `yaml:"ssm-parameter"`
}

func (cfg *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Config
	a := struct {
		alias                     `yaml:",inline"`
		LogLevel                  string `yaml:"log-level"`
		BuildStatusContext        string `yaml:"build-status-context"`
		ErrorNotificationTemplate string `yaml:"error-notification-template"`
	}{}
	if err := unmarshal(&a); err != nil {
		return err
	}
	*cfg = Config(a.alias)
	if a.LogLevel != "" {
		lvl, err := logrus.ParseLevel(a.LogLevel)
		if err != nil {
			return fmt.Errorf("log-level is invalid (%s): %w", a.LogLevel, err)
		}
		cfg.LogLevel = lvl
	}
	if a.BuildStatusContext != "" {
		tpl, err := templ.Compile(a.BuildStatusContext)
		if err != nil {
			return fmt.Errorf("parse build-status-context as template (%s): %w", a.BuildStatusContext, err)
		}
		cfg.BuildStatusContext = tpl
	}
	if a.ErrorNotificationTemplate != "" {
		tpl, err := templ.Compile(a.ErrorNotificationTemplate)
		if err != nil {
			return fmt.Errorf("parse error-notification-template as template (%s): %w", a.ErrorNotificationTemplate, err)
		}
		cfg.ErrorNotificationTemplate = tpl
	}
	return nil
}

type SSMParameter struct {
	ParameterName ParameterName `yaml:"parameter-name"`
}

type ParameterName struct {
	GitHubToken   string `yaml:"github-token"`
	WebhookSecret string `yaml:"webhook-secret"`
}

type Repository struct {
	Name      string
	Hooks     []Hook
	CodeBuild CodeBuild `yaml:"codebuild"`
}

type CodeBuild struct {
	ProjectName string `yaml:"project-name"`
}

type Hook struct {
	If     *vm.Program `yaml:"-"`
	Config string
}

func (hook *Hook) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Hook
	a := struct {
		alias `yaml:",inline"`
		If    string
	}{}
	if err := unmarshal(&a); err != nil {
		return err
	}
	*hook = Hook(a.alias)
	if a.If != "" {
		prog, err := expr.Compile(a.If, expr.AsBool())
		if err != nil {
			return fmt.Errorf("compile an expression: %s: %w", a.If, err)
		}
		hook.If = prog
	}
	return nil
}
