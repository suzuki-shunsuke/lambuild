package generator

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"gopkg.in/yaml.v2"
)

func handleBuild(data *domain.Data, buildspec bspec.Buildspec) (domain.BuildInput, error) {
	buildInput := domain.BuildInput{
		BatchBuild: &codebuild.StartBuildBatchInput{},
	}
	items := buildspec.Lambuild.Items
	if len(items) == 0 {
		items = []interface{}{
			map[string]interface{}{},
		}
	}
	builds := make([]*codebuild.StartBuildInput, 0, len(items))

	for range items {
		build := &codebuild.StartBuildInput{}
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
		build.EnvironmentVariablesOverride = envs

		if !buildspec.Lambuild.BuildStatusContext.Empty() {
			s, err := buildspec.Lambuild.BuildStatusContext.Execute(data.Convert())
			if err != nil {
				return buildInput, fmt.Errorf("render a build status context: %w", err)
			}
			build.BuildStatusConfigOverride = &codebuild.BuildStatusConfig{
				Context: aws.String(s),
			}
		}

		if buildspec.Lambuild.Image != "" {
			build.ImageOverride = aws.String(buildspec.Lambuild.Image)
		}

		if buildspec.Lambuild.ComputeType != "" {
			build.ComputeTypeOverride = aws.String(buildspec.Lambuild.ComputeType)
		}

		if buildspec.Lambuild.EnvironmentType != "" {
			build.EnvironmentTypeOverride = aws.String(buildspec.Lambuild.EnvironmentType)
		}

		if buildspec.Lambuild.DebugSession != nil {
			build.DebugSessionEnabled = buildspec.Lambuild.DebugSession
		}

		if buildspec.Lambuild.GitCloneDepth != nil {
			build.GitCloneDepthOverride = buildspec.Lambuild.GitCloneDepth
		}

		if buildspec.Lambuild.PrivilegedMode != nil {
			build.PrivilegedModeOverride = buildspec.Lambuild.PrivilegedMode
		}

		buildspec.Lambuild = bspec.Lambuild{}
		builtContent, err := yaml.Marshal(&buildspec)
		if err != nil {
			return buildInput, fmt.Errorf("marshal a buildspec: %w", err)
		}
		build.BuildspecOverride = aws.String(string(builtContent))
		builds = append(builds, build)
	}
	buildInput.Builds = builds
	return buildInput, nil
}
