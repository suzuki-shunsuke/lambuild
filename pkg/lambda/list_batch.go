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

func (handler *Handler) handleList(ctx context.Context, logE *logrus.Entry, data *Data, buildspec bspec.Buildspec, repo config.Repository) error {
	listElems, err := handler.extractBuildList(data, buildspec.Batch.BuildList)
	if err != nil {
		return err
	}
	if len(listElems) == 0 {
		logE.Info("no list element is run")
		return nil
	}

	if len(listElems) == 1 {
		input, err := handler.getListBuildInput(data, repo, listElems[0])
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
	buildspec.Batch.BuildList = listElems
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

func (handler *Handler) getListBuildInput(data *Data, repo config.Repository, listElem bspec.ListElement) (*codebuild.StartBuildInput, error) {
	input := &codebuild.StartBuildInput{
		BuildspecOverride: aws.String(listElem.Buildspec),
		ProjectName:       aws.String(repo.CodeBuild.ProjectName),
		SourceVersion:     aws.String(data.SHA),
	}

	if listElem.Env.ComputeType != "" {
		input.ComputeTypeOverride = aws.String(listElem.Env.ComputeType)
	}

	if listElem.Env.Image != "" {
		input.ImageOverride = aws.String(listElem.Env.Image)
	}

	if listElem.Env.PrivilegedMode {
		input.PrivilegedModeOverride = aws.Bool(true)
	}

	if err := handler.setBuildStatusContext(data, input); err != nil {
		return nil, err
	}

	if listElem.DebugSession {
		input.DebugSessionEnabled = aws.Bool(true)
	}

	if len(listElem.Env.Variables) == 0 {
		listElem.Env.Variables = make(map[string]string, len(data.Lambuild.Env.Variables))
	}
	for k, prog := range data.Lambuild.Env.Variables {
		if _, ok := listElem.Env.Variables[k]; ok {
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
		listElem.Env.Variables[k] = s
	}
	envs := make([]*codebuild.EnvironmentVariable, 0, len(listElem.Env.Variables))
	for k, v := range listElem.Env.Variables {
		envs = append(envs, &codebuild.EnvironmentVariable{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	input.EnvironmentVariablesOverride = envs

	return input, nil
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
