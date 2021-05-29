package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

type Config struct {
	Region                    string
	Repositories              []Repository
	LogLevel                  LogLevel          `yaml:"log-level"`
	BuildStatusContext        template.Template `yaml:"build-status-context"`
	ErrorNotificationTemplate template.Template `yaml:"error-notification-template"`
	SSMParameter              SSMParameter      `yaml:"ssm-parameter"`
}

type LogLevel struct {
	level logrus.Level
}

func (logLevel *LogLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	lvl, err := logrus.ParseLevel(s)
	if err != nil {
		return fmt.Errorf("log-level is invalid (%s): %w", s, err)
	}
	logLevel.level = lvl
	return nil
}

func (logLevel *LogLevel) Get() logrus.Level {
	return logLevel.level
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
	If          expr.Bool
	Config      string
	ServiceRole string `yaml:"service-role"`
}
