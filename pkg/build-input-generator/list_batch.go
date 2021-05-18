package generator

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func handleList(buildInput *domain.BuildInput, logE *logrus.Entry, buildStatusContext template.Template, data *domain.Data, buildspec bspec.Buildspec) error {
	listElems, err := extractBuildList(data, buildspec.Batch.BuildList)
	if err != nil {
		return err
	}
	if len(listElems) == 0 {
		buildInput.Empty = true
		logE.Info("no list element is run")
		return nil
	}

	if len(listElems) == 1 { //nolint:nestif
		elem := listElems[0]
		build := &codebuild.StartBuildInput{
			BuildspecOverride: aws.String(elem.Buildspec),
		}
		if err := setListBuildInput(build, buildStatusContext, data, buildspec.Lambuild, elem); err != nil {
			return fmt.Errorf("set a codebuild.StartBuildInput: %w", err)
		}
		if elem.Buildspec == "" {
			buildspec.Batch = bspec.Batch{}
			s, err := buildspec.ToYAML(data.Convert())
			if err != nil {
				return fmt.Errorf("render a buildspec: %w", err)
			}
			build.BuildspecOverride = aws.String(string(s))
		} else {
			build.BuildspecOverride = aws.String(elem.Buildspec)
		}
		buildInput.Builds = []*codebuild.StartBuildInput{build}
		return nil
	}

	buildInput.Batched = true
	buildspec.Batch.BuildList = listElems
	if err := setBatchBuildInput(buildInput.BatchBuild, buildspec, data); err != nil {
		return fmt.Errorf("set codebuild.StartBuildBatchInput: %w", err)
	}
	return nil
}

func setListBuildInput(input *codebuild.StartBuildInput, contx template.Template, data *domain.Data, lambuild bspec.Lambuild, elem bspec.ListElement) error {
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

	if err := setEnvsToStartBuildInput(input, data, lambuild, elem.Env.Variables); err != nil {
		return fmt.Errorf("set EnvironmentVariablesOverride: %w", err)
	}

	return nil
}

func extractBuildList(data *domain.Data, allElems []bspec.ListElement) ([]bspec.ListElement, error) {
	listElems := []bspec.ListElement{}
	for _, listElem := range allElems {
		if listElem.If.Empty() {
			listElems = append(listElems, listElem)
			continue
		}
		f, err := listElem.If.Run(data.Convert())
		if err != nil {
			return nil, fmt.Errorf("evaluate an expression: %w", err)
		}
		if !f {
			continue
		}
		listElem.If = expr.Bool{}
		listElems = append(listElems, listElem)
	}
	return listElems, nil
}
