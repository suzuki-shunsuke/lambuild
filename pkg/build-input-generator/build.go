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
		items = []bspec.Item{
			{},
		}
	}
	builds := make([]*codebuild.StartBuildInput, 0, len(items))

	for _, item := range items {
		build, err := handleBuildItem(data, buildspec, item)
		if err != nil {
			return buildInput, err
		}
		if build.BuildspecOverride != nil {
			builds = append(builds, &build)
		}
	}
	buildInput.Builds = builds
	return buildInput, nil
}

func handleBuildItem(data *domain.Data, buildspec bspec.Buildspec, item bspec.Item) (codebuild.StartBuildInput, error) {
	build := codebuild.StartBuildInput{}
	param := data.Convert()
	param["item"] = item.Param

	if !item.If.Empty() {
		f, err := item.If.Run(param)
		if err != nil {
			return build, fmt.Errorf("evaluate item.If: %w", err)
		}
		if !f {
			return build, nil
		}
	}

	envMap := map[string]string{}
	for k, prog := range buildspec.Lambuild.Env.Variables {
		s, err := prog.Run(param)
		if err != nil {
			return build, fmt.Errorf("evaluate an expression: %w", err)
		}
		envMap[k] = s
	}

	for k, prog := range item.Env.Variables {
		s, err := prog.Run(param)
		if err != nil {
			return build, fmt.Errorf("evaluate an expression: %w", err)
		}
		envMap[k] = s
	}

	envs := make([]*codebuild.EnvironmentVariable, 0, len(envMap))
	for k, v := range envMap {
		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}
	build.EnvironmentVariablesOverride = envs

	if !item.BuildStatusContext.Empty() {
		s, err := item.BuildStatusContext.Execute(param)
		if err != nil {
			return build, fmt.Errorf("render a build status context: %w", err)
		}
		build.BuildStatusConfigOverride = &codebuild.BuildStatusConfig{
			Context: aws.String(s),
		}
	} else if !buildspec.Lambuild.BuildStatusContext.Empty() {
		s, err := buildspec.Lambuild.BuildStatusContext.Execute(param)
		if err != nil {
			return build, fmt.Errorf("render a build status context: %w", err)
		}
		build.BuildStatusConfigOverride = &codebuild.BuildStatusConfig{
			Context: aws.String(s),
		}
	}

	if item.Image != "" {
		build.ImageOverride = aws.String(item.Image)
	} else if buildspec.Lambuild.Image != "" {
		build.ImageOverride = aws.String(buildspec.Lambuild.Image)
	}

	if item.ComputeType != "" {
		build.ComputeTypeOverride = aws.String(item.ComputeType)
	} else if buildspec.Lambuild.ComputeType != "" {
		build.ComputeTypeOverride = aws.String(buildspec.Lambuild.ComputeType)
	}

	if item.EnvironmentType != "" {
		build.EnvironmentTypeOverride = aws.String(item.EnvironmentType)
	} else if buildspec.Lambuild.EnvironmentType != "" {
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

	m, err := buildspec.Filter(param)
	if err != nil {
		return build, fmt.Errorf("filter commands from buildspec: %w", err)
	}

	builtContent, err := yaml.Marshal(m)
	if err != nil {
		return build, fmt.Errorf("marshal a buildspec: %w", err)
	}
	build.BuildspecOverride = aws.String(string(builtContent))
	return build, nil
}
