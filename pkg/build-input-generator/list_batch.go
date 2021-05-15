package generator

import (
	"errors"
	"fmt"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"gopkg.in/yaml.v2"
)

func handleList(buildInput *domain.BuildInput, logE *logrus.Entry, buildStatusContext *template.Template, data *domain.Data, buildspec bspec.Buildspec) error {
	listElems, err := extractBuildList(data, buildspec.Batch.BuildList)
	if err != nil {
		return err
	}
	if len(listElems) == 0 {
		buildInput.Empty = true
		logE.Info("no list element is run")
		return nil
	}

	if len(listElems) == 1 {
		elem := listElems[0]
		buildInput.Build.BuildspecOverride = aws.String(elem.Buildspec)
		if err := setListBuildInput(buildInput.Build, buildStatusContext, data, buildspec.Lambuild, elem); err != nil {
			return fmt.Errorf("set a codebuild.StartBuildInput: %w", err)
		}
		return nil
	}

	buildInput.Batched = true
	buildspec.Batch.BuildList = listElems
	if err := setBatchBuildInput(buildInput.BatchBuild, buildspec, data); err != nil {
		return fmt.Errorf("set codebuild.StartBuildBatchInput: %w", err)
	}
	return nil
}

func getLambuildEnvVars(data *domain.Data, lambuild bspec.Lambuild) ([]*codebuild.EnvironmentVariable, error) {
	envs := make([]*codebuild.EnvironmentVariable, 0, len(lambuild.Env.Variables))
	for k, prog := range lambuild.Env.Variables {
		a, err := domain.RunExpr(prog, data)
		if err != nil {
			return nil, fmt.Errorf("evaluate an expression: %w", err)
		}
		s, ok := a.(string)
		if !ok {
			return nil, errors.New("the evaluated result must be string: lambuild.env." + k)
		}
		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String(k),
			Value: aws.String(s),
		})
	}
	return envs, nil
}

func setBatchBuildInput(input *codebuild.StartBuildBatchInput, buildspec bspec.Buildspec, data *domain.Data) error {
	envs, err := getLambuildEnvVars(data, buildspec.Lambuild)
	if err != nil {
		return err
	}
	input.EnvironmentVariablesOverride = envs

	buildspec.Lambuild = bspec.Lambuild{}
	builtContent, err := yaml.Marshal(&buildspec)
	if err != nil {
		return fmt.Errorf("marshal a buildspec: %w", err)
	}
	input.BuildspecOverride = aws.String(string(builtContent))

	return nil
}

func setListBuildInput(input *codebuild.StartBuildInput, contx *template.Template, data *domain.Data, lambuild bspec.Lambuild, elem bspec.ListElement) error { //nolint:dupl
	if elem.Env.ComputeType != "" {
		input.ComputeTypeOverride = aws.String(elem.Env.ComputeType)
	}

	if elem.DebugSession {
		input.DebugSessionEnabled = aws.Bool(true)
	}

	if elem.Env.Image != "" {
		input.ImageOverride = aws.String(elem.Env.Image)
	}

	if elem.Env.PrivilegedMode {
		input.PrivilegedModeOverride = aws.Bool(true)
	}

	if err := setBuildStatusContext(contx, data, input); err != nil {
		return err
	}

	envMap := make(map[string]string, len(elem.Env.Variables)+len(lambuild.Env.Variables))
	for k, prog := range lambuild.Env.Variables {
		a, err := domain.RunExpr(prog, data)
		if err != nil {
			return fmt.Errorf("evaluate an expression: %w", err)
		}
		s, ok := a.(string)
		if !ok {
			return errors.New("the evaluated result must be string: lambuild.env." + k)
		}
		envMap[k] = s
	}
	for k, v := range elem.Env.Variables {
		envMap[k] = v
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

func extractBuildList(data *domain.Data, allElems []bspec.ListElement) ([]bspec.ListElement, error) {
	listElems := []bspec.ListElement{}
	for _, listElem := range allElems {
		if listElem.If == nil {
			listElems = append(listElems, listElem)
			continue
		}
		f, err := domain.RunExpr(listElem.If, data)
		if err != nil {
			return nil, fmt.Errorf("evaluate an expression: %w", err)
		}
		if !f.(bool) {
			continue
		}
		listElem.If = nil
		listElems = append(listElems, listElem)
	}
	return listElems, nil
}
