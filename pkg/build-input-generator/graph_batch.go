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

func handleGraph(buildStatusContext template.Template, buildInput *domain.BuildInput, logE *logrus.Entry, data *domain.Data, buildspec bspec.Buildspec) error {
	elems, err := extractGraph(logE, data, buildspec.Batch.BuildGraph)
	if err != nil {
		return err
	}
	if len(elems) == 0 {
		logE.Info("no graph element is run")
		buildInput.Empty = true
		return nil
	}

	if len(elems) == 1 { //nolint:nestif
		elem := elems[0]
		build := &codebuild.StartBuildInput{}
		if err := setGraphBuildInput(build, buildStatusContext, data, buildspec.Lambuild, elem); err != nil {
			return fmt.Errorf("set codebuild.StartBuildInput: %w", err)
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
	buildspec.Batch.BuildGraph = elems
	if err := setBatchBuildInput(buildInput.BatchBuild, buildspec, data); err != nil {
		return fmt.Errorf("set codebuild.StartBuildBatchInput: %w", err)
	}
	return nil
}

func extractGraphByDependency(identifiers map[string]bspec.GraphElement, logE *logrus.Entry) {
	for {
		removed := false
		for identifier, elem := range identifiers {
			for _, dep := range elem.DependOn {
				if _, ok := identifiers[dep]; !ok {
					logE.WithFields(logrus.Fields{
						"build_identifier":     identifier,
						"dependent_identifier": dep,
					}).Info("a build isn't run because a dependent build isn't run")
					delete(identifiers, identifier)
					removed = true
					break
				}
			}
		}
		if !removed {
			break
		}
	}
}

func extractGraphByIf(data *domain.Data, allElems []bspec.GraphElement, identifiers map[string]bspec.GraphElement) error {
	for _, elem := range allElems {
		if elem.If.Empty() {
			identifiers[elem.Identifier] = elem
			continue
		}
		f, err := elem.If.Run(data.Convert())
		if err != nil {
			return fmt.Errorf("evaluate an expression: %w", err)
		}
		if !f {
			continue
		}
		elem.If = expr.Bool{}
		identifiers[elem.Identifier] = elem
	}
	return nil
}

func extractGraph(logE *logrus.Entry, data *domain.Data, allElems []bspec.GraphElement) ([]bspec.GraphElement, error) {
	identifiers := make(map[string]bspec.GraphElement, len(allElems))
	if err := extractGraphByIf(data, allElems, identifiers); err != nil {
		return nil, err
	}
	extractGraphByDependency(identifiers, logE)

	if len(identifiers) == 0 {
		return nil, nil
	}

	elems := make([]bspec.GraphElement, 0, len(identifiers))
	for _, elem := range identifiers {
		elems = append(elems, elem)
	}
	return elems, nil
}

func setGraphBuildInput(input *codebuild.StartBuildInput, buildStatusContext template.Template, data *domain.Data, lambuild bspec.Lambuild, elem bspec.GraphElement) error {
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

	if err := setBuildStatusContext(buildStatusContext, data, input); err != nil {
		return err
	}

	if err := setEnvsToStartBuildInput(input, data, lambuild, elem.Env.Variables); err != nil {
		return fmt.Errorf("set EnvironmentVariablesOverride: %w", err)
	}

	return nil
}
