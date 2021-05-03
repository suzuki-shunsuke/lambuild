package lambda

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"gopkg.in/yaml.v2"
)

func (handler *Handler) handleMatrix(buildInput *BuildInput, logE *logrus.Entry, data *Data, buildspec bspec.Buildspec, repo config.Repository) error {
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
		buildInput.Empty = true
		return nil
	}

	if len(buildspecs) > 1 || len(images) > 1 || len(computeTypes) > 1 || getSizeOfEnvVars(envVars) > 1 {
		// batch build
		buildInput.Batched = true
		if err := handler.setMatrixBatchBuildInput(buildInput.BatchBuild, data, buildspec, repo, buildspecs, images, computeTypes, envVars); err != nil {
			return err
		}
		return nil
	}

	// build
	if err := handler.setMatrixBuildInput(data, repo, buildspecs, images, computeTypes, envVars, buildInput.Build); err != nil {
		return err
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
		a := bs.(bspec.ExprElem)
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

func (handler *Handler) setMatrixBatchBuildInput(input *codebuild.StartBuildBatchInput, data *Data, buildspec bspec.Buildspec, repo config.Repository, buildspecs, images, computeTypes bspec.ExprList, envVars map[string]bspec.ExprList) error {
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

	input.BuildspecOverride = aws.String(string(builtContent))
	input.ProjectName = aws.String(repo.CodeBuild.ProjectName)
	input.SourceVersion = aws.String(data.SHA)
	input.EnvironmentVariablesOverride = envs
	return nil
}

func (handler *Handler) setMatrixBuildInput(data *Data, repo config.Repository, buildspecs, images, computeTypes bspec.ExprList, envVars map[string]bspec.ExprList, input *codebuild.StartBuildInput) error {
	input.ProjectName = aws.String(repo.CodeBuild.ProjectName)
	input.SourceVersion = aws.String(data.SHA)
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
	if getSizeOfEnvVars(envVars) == 0 {
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
		input.EnvironmentVariablesOverride = envs
	} else {
		list := make([]*codebuild.EnvironmentVariable, 0, len(envVars))
		for k, v := range envVars {
			list = append(list, &codebuild.EnvironmentVariable{
				Name:  aws.String(k),
				Value: aws.String(v[0].(string)),
			})
		}

		for k, prog := range data.Lambuild.Env.Variables {
			if _, ok := envVars[k]; ok {
				continue
			}
			a, err := runExpr(prog, data)
			if err != nil {
				return err
			}
			s, ok := a.(string)
			if !ok {
				return errors.New("the evaluated result must be string: lambuild.env." + k)
			}
			list = append(list, &codebuild.EnvironmentVariable{
				Name:  aws.String(k),
				Value: aws.String(s),
			})
		}

		input.EnvironmentVariablesOverride = list
	}

	return nil
}
