package generator

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func GenerateInput(logE *logrus.Entry, buildStatusContext template.Template, data *domain.Data, buildspec bspec.Buildspec, repo config.Repository) (domain.BuildInput, error) {
	buildInput := domain.BuildInput{
		BatchBuild: &codebuild.StartBuildBatchInput{},
	}

	if !buildspec.Lambuild.If.Empty() {
		f, err := buildspec.Lambuild.If.Run(data.Convert())
		if err != nil {
			return buildInput, fmt.Errorf("evaluate buildspec.Lambuild.If: %w", err)
		}
		if !f {
			return domain.BuildInput{
				Empty: true,
			}, nil
		}
	}

	return handleBuild(data, buildspec)
}

func setEnvsToStartBuildInput(input *codebuild.StartBuildInput, data *domain.Data, lambuild bspec.Lambuild, envVars map[string]string) error {
	envMap := make(map[string]string, len(envVars)+len(lambuild.Env.Variables))
	for k, prog := range lambuild.Env.Variables {
		s, err := prog.Run(data.Convert())
		if err != nil {
			return fmt.Errorf("evaluate an expression: %w", err)
		}
		envMap[k] = s
	}
	for k, v := range envVars {
		envMap[k] = v
	}

	if len(envMap) == 0 {
		return nil
	}

	envs := make([]*codebuild.EnvironmentVariable, 0, len(envMap))
	for k, v := range envMap {
		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}
	input.EnvironmentVariablesOverride = envs
	return nil
}

func getLambuildEnvVars(data *domain.Data, lambuild bspec.Lambuild) ([]*codebuild.EnvironmentVariable, error) {
	envs := make([]*codebuild.EnvironmentVariable, 0, len(lambuild.Env.Variables))
	for k, prog := range lambuild.Env.Variables {
		s, err := prog.Run(data.Convert())
		if err != nil {
			return nil, fmt.Errorf("evaluate an expression: %w", err)
		}
		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String(k),
			Value: aws.String(s),
		})
	}
	return envs, nil
}
