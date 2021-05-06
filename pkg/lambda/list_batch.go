package lambda

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"gopkg.in/yaml.v2"
)

func (handler *Handler) handleList(buildInput *BuildInput, logE *logrus.Entry, data *Data, buildspec bspec.Buildspec) error {
	listElems, err := handler.extractBuildList(data, buildspec.Batch.BuildList)
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
		if err := handler.setListBuildInput(buildInput.Build, data, elem); err != nil {
			return fmt.Errorf("set a codebuild.StartBuildInput: %w", err)
		}
		return nil
	}

	buildInput.Batched = true
	buildspec.Batch.BuildList = listElems
	if err := handler.setBatchBuildInput(buildInput.BatchBuild, buildspec, data); err != nil {
		return fmt.Errorf("set codebuild.StartBuildBatchInput: %w", err)
	}
	return nil
}

func (handler *Handler) getLambuildEnvVars(data *Data) ([]*codebuild.EnvironmentVariable, error) {
	envs := make([]*codebuild.EnvironmentVariable, 0, len(data.Lambuild.Env.Variables))
	for k, prog := range data.Lambuild.Env.Variables {
		a, err := runExpr(prog, data)
		if err != nil {
			return nil, err
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

func (handler *Handler) setBatchBuildInput(input *codebuild.StartBuildBatchInput, buildspec bspec.Buildspec, data *Data) error {
	envs, err := handler.getLambuildEnvVars(data)
	if err != nil {
		return err
	}
	input.EnvironmentVariablesOverride = envs

	builtContent, err := yaml.Marshal(&buildspec)
	if err != nil {
		return fmt.Errorf("marshal a buildspec: %w", err)
	}
	input.BuildspecOverride = aws.String(string(builtContent))

	return nil
}

func (handler *Handler) setListBuildInput(input *codebuild.StartBuildInput, data *Data, elem bspec.ListElement) error { //nolint:dupl
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

	if err := handler.setBuildStatusContext(data, input); err != nil {
		return err
	}

	envMap := make(map[string]string, len(elem.Env.Variables)+len(data.Lambuild.Env.Variables))
	for k, prog := range data.Lambuild.Env.Variables {
		a, err := runExpr(prog, data)
		if err != nil {
			return err
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

func (handler *Handler) extractBuildList(data *Data, allElems []bspec.ListElement) ([]bspec.ListElement, error) {
	listElems := []bspec.ListElement{}
	for _, listElem := range allElems {
		if listElem.If == nil {
			listElems = append(listElems, listElem)
			continue
		}
		f, err := runExpr(listElem.If, data)
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
