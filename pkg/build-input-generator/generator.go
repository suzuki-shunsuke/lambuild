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
	"gopkg.in/yaml.v2"
)

func GenerateInput(logE *logrus.Entry, buildStatusContext template.Template, data *domain.Data, buildspec bspec.Buildspec, repo config.Repository) (domain.BuildInput, error) {
	buildInput := domain.BuildInput{
		Build: &codebuild.StartBuildInput{
			ProjectName:   aws.String(repo.CodeBuild.ProjectName),
			SourceVersion: aws.String(data.SHA),
		},
		BatchBuild: &codebuild.StartBuildBatchInput{
			ProjectName:   aws.String(repo.CodeBuild.ProjectName),
			SourceVersion: aws.String(data.SHA),
		},
	}

	if len(buildspec.Batch.BuildGraph) != 0 {
		logE.Debug("handling build-graph")
		if err := handleGraph(buildStatusContext, &buildInput, logE, data, buildspec); err != nil {
			return buildInput, err
		}
	}

	if len(buildspec.Batch.BuildList) != 0 {
		logE.Debug("handling build-list")
		if err := handleList(&buildInput, logE, buildStatusContext, data, buildspec); err != nil {
			return buildInput, err
		}
	}

	if !buildspec.Batch.BuildMatrix.Empty() {
		logE.Debug("handling build-matrix")
		if err := handleMatrix(&buildInput, logE, buildStatusContext, data, buildspec); err != nil {
			return buildInput, err
		}
	}

	envs := make([]*codebuild.EnvironmentVariable, 0, len(buildspec.Lambuild.Env.Variables))
	for k, prog := range buildspec.Lambuild.Env.Variables {
		s, err := prog.Run(data.Convert())
		if err != nil {
			return buildInput, fmt.Errorf("evaluate an expression: %w", err)
		}

		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String(k),
			Value: aws.String(s),
		})
	}
	buildInput.Build.EnvironmentVariablesOverride = envs
	buildspec.Lambuild = bspec.Lambuild{}
	builtContent, err := yaml.Marshal(&buildspec)
	if err != nil {
		return buildInput, fmt.Errorf("marshal a buildspec: %w", err)
	}
	buildInput.Build.BuildspecOverride = aws.String(string(builtContent))

	return buildInput, nil
}
