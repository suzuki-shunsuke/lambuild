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
	SecretsManager            SecretsManager    `yaml:"secrets-manager"`
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

type Repository struct {
	Name      string
	Hooks     []Hook
	CodeBuild CodeBuild `yaml:"codebuild"`
}

type CodeBuild struct {
	ProjectName   string `yaml:"project-name"`
	AssumeRoleARN string `yaml:"assume-role-arn"`
}

type Hook struct {
	If            expr.Bool
	Config        string
	ServiceRole   string `yaml:"service-role"`
	ProjectName   string `yaml:"project-name"`
	AssumeRoleARN string `yaml:"assume-role-arn"`
}

type SecretsManager struct {
	SecretID  string `yaml:"secret-id"`
	VersionID string `yaml:"version-id"`
}
