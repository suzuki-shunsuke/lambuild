package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"gopkg.in/yaml.v2"
)

func (handler *Handler) handleGraph(ctx context.Context, logE *logrus.Entry, data *Data, buildspec bspec.Buildspec, repo config.Repository) error {
	elems, err := handler.extractGraph(logE, data, buildspec.Batch.BuildGraph)
	if err != nil {
		return err
	}
	if len(elems) == 0 {
		logE.Info("no graph element is run")
		return nil
	}

	if len(elems) == 1 {
		input, err := handler.getGraphBuildInput(data, repo, elems[0])
		if err != nil {
			return err
		}

		buildOut, err := handler.CodeBuild.StartBuildWithContext(ctx, input)
		if err != nil {
			return fmt.Errorf("start a batch build: %w", err)
		}
		logE.WithFields(logrus.Fields{
			"build_arn": *buildOut.Build.Arn,
		}).Info("start a build")
		return nil
	}
	buildspec.Batch.BuildGraph = elems
	builtContent, err := yaml.Marshal(buildspec)
	if err != nil {
		return err
	}

	envs := make([]*codebuild.EnvironmentVariable, 0, len(data.Lambuild.Env.Variables))
	for k, prog := range data.Lambuild.Env.Variables {
		a, err := runExpr(prog, data)
		if err != nil {
			return err
		}
		s, ok := a.(string)
		if !ok {
			return errors.New("the evaluated result must be string: lambuild.env." + k)
		}
		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String(k),
			Value: aws.String(s),
		})
	}

	buildOut, err := handler.CodeBuild.StartBuildBatchWithContext(ctx, &codebuild.StartBuildBatchInput{
		BuildspecOverride:            aws.String(string(builtContent)),
		ProjectName:                  aws.String(repo.CodeBuild.ProjectName),
		SourceVersion:                aws.String(data.SHA),
		EnvironmentVariablesOverride: envs,
	})
	if err != nil {
		return fmt.Errorf("start a batch build: %w", err)
	}
	logE.WithFields(logrus.Fields{
		"build_arn": *buildOut.BuildBatch.Arn,
	}).Info("start a batch build")
	return nil
}

func (handler *Handler) extractGraph(logE *logrus.Entry, data *Data, allElems []bspec.GraphElement) ([]bspec.GraphElement, error) {
	identifiers := make(map[string]bspec.GraphElement, len(allElems))
	for _, elem := range allElems {
		if elem.If == nil {
			identifiers[elem.Identifier] = elem
			continue
		}
		f, err := runExpr(elem.If, data)
		if err != nil {
			return nil, fmt.Errorf("evaluate an expression: %w", err)
		}
		if !f.(bool) {
			continue
		}
		elem.If = nil
		identifiers[elem.Identifier] = elem
	}
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

	if len(identifiers) == 0 {
		return nil, nil
	}

	elems := make([]bspec.GraphElement, 0, len(identifiers))
	for _, elem := range identifiers {
		elems = append(elems, elem)
	}
	return elems, nil
}

func (handler *Handler) getGraphBuildInput(data *Data, repo config.Repository, graphElem bspec.GraphElement) (*codebuild.StartBuildInput, error) {
	input := &codebuild.StartBuildInput{
		BuildspecOverride: aws.String(graphElem.Buildspec),
		ProjectName:       aws.String(repo.CodeBuild.ProjectName),
		SourceVersion:     aws.String(data.SHA),
	}

	if graphElem.DebugSession {
		input.DebugSessionEnabled = aws.Bool(true)
	}

	if graphElem.Env.ComputeType != "" {
		input.ComputeTypeOverride = aws.String(graphElem.Env.ComputeType)
	}

	if graphElem.Env.Image != "" {
		input.ImageOverride = aws.String(graphElem.Env.Image)
	}

	if graphElem.Env.PrivilegedMode {
		input.PrivilegedModeOverride = aws.Bool(true)
	}

	if err := handler.setBuildStatusContext(data, input); err != nil {
		return nil, err
	}

	if graphElem.Env.Variables == nil {
		// initialize a nil map
		graphElem.Env.Variables = make(map[string]string, len(data.Lambuild.Env.Variables))
	}

	for k, prog := range data.Lambuild.Env.Variables {
		if _, ok := graphElem.Env.Variables[k]; ok {
			continue
		}
		a, err := runExpr(prog, data)
		if err != nil {
			return nil, err
		}
		s, ok := a.(string)
		if !ok {
			return nil, errors.New("the evaluated result must be string: lambuild.env." + k)
		}
		graphElem.Env.Variables[k] = s
	}
	envs := make([]*codebuild.EnvironmentVariable, 0, len(graphElem.Env.Variables))
	for k, v := range graphElem.Env.Variables {
		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}
	input.EnvironmentVariablesOverride = envs
	return input, nil
}
