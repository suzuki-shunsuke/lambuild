package lambda

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
)

func (handler *Handler) handleMatrix(buildInput *BuildInput, logE *logrus.Entry, data *Data, buildspec bspec.Buildspec) error {
	buildspecs, err := handler.filterExprList(data, buildspec.Batch.BuildMatrix.Dynamic.Buildspec)
	if err != nil {
		return fmt.Errorf("filter buildspecs: %w", err)
	}
	buildspec.Batch.BuildMatrix.Dynamic.Buildspec = buildspecs

	images, err := handler.filterExprList(data, buildspec.Batch.BuildMatrix.Dynamic.Env.Image)
	if err != nil {
		return fmt.Errorf("filter images: %w", err)
	}
	buildspec.Batch.BuildMatrix.Dynamic.Env.Image = images

	computeTypes, err := handler.filterExprList(data, buildspec.Batch.BuildMatrix.Dynamic.Env.ComputeType)
	if err != nil {
		return fmt.Errorf("filter compute-type: %w", err)
	}
	buildspec.Batch.BuildMatrix.Dynamic.Env.ComputeType = computeTypes

	envVars := make(map[string]bspec.ExprList, len(buildspec.Batch.BuildMatrix.Dynamic.Env.Variables))
	for k, v := range buildspec.Batch.BuildMatrix.Dynamic.Env.Variables {
		vars, err := handler.filterExprList(data, v)
		if err != nil {
			return fmt.Errorf("filter env.variables: %w", err)
		}
		envVars[k] = vars
	}
	buildspec.Batch.BuildMatrix.Dynamic.Env.Variables = envVars

	if len(buildspecs) == 0 && len(images) == 0 && len(computeTypes) == 0 && getSizeOfEnvVars(envVars) == 0 {
		logE.Info("no matrix element is run")
		buildInput.Empty = true
		return nil
	}

	if len(buildspecs) > 1 || len(images) > 1 || len(computeTypes) > 1 || getSizeOfEnvVars(envVars) > 1 {
		// batch build
		buildInput.Batched = true
		if err := handler.setBatchBuildInput(buildInput.BatchBuild, buildspec, data); err != nil {
			return fmt.Errorf("set codebuild.StartBuildBatchInput: %w", err)
		}
		return nil
	}

	// build
	if err := handler.setMatrixBuildInput(data, buildspecs, images, computeTypes, envVars, buildInput.Build); err != nil {
		return fmt.Errorf("set codebuild.StartBuildInput: %w", err)
	}

	return nil
}

func (handler *Handler) filterExprList(data *Data, src bspec.ExprList) (bspec.ExprList, error) {
	list := bspec.ExprList{}
	for _, bs := range src {
		s, ok := bs.(string)
		if ok {
			list = append(list, s)
			continue
		}
		a := bs.(bspec.ExprElem) //nolint:forcetypeassert
		if a.If == nil {
			list = append(list, a.Value)
			continue
		}
		f, err := runExpr(a.If, data)
		if err != nil {
			return nil, fmt.Errorf("evaluate an expression: %w", err)
		}
		if f.(bool) {
			list = append(list, a.Value)
		}
	}
	return list, nil
}

func getSizeOfEnvVars(m map[string]bspec.ExprList) int {
	size := 1
	for _, v := range m {
		size *= len(v)
	}
	return size
}

func (handler *Handler) setMatrixBuildInput(data *Data, buildspecs, images, computeTypes bspec.ExprList, envVars map[string]bspec.ExprList, input *codebuild.StartBuildInput) error {
	if err := handler.setBuildStatusContext(data, input); err != nil {
		return err
	}

	if len(buildspecs) != 0 {
		input.BuildspecOverride = aws.String(buildspecs[0].(string))
	}
	if len(images) != 0 {
		input.ImageOverride = aws.String(images[0].(string))
	}
	if len(computeTypes) != 0 {
		input.ComputeTypeOverride = aws.String(computeTypes[0].(string))
	}

	envMap := make(map[string]string, len(data.Lambuild.Env.Variables))
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
	if getSizeOfEnvVars(envVars) != 0 {
		for k, v := range envVars {
			envMap[k] = v[0].(string) //nolint:forcetypeassert
		}
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
