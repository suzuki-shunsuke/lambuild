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
)

func handleMatrix(buildInput *domain.BuildInput, logE *logrus.Entry, buildStatusContext *template.Template, data *domain.Data, buildspec bspec.Buildspec) error { //nolint:gocognit
	dynamic := buildspec.Batch.BuildMatrix.Dynamic
	if len(dynamic.Buildspec) != 0 {
		buildspecs, err := filterExprList(data, dynamic.Buildspec)
		if err != nil {
			return fmt.Errorf("filter buildspecs: %w", err)
		}
		if len(buildspecs) == 0 {
			buildInput.Empty = true
			logE.Info("all buildspec are ignored, so build isn't run")
			return nil
		}
		dynamic.Buildspec = buildspecs
	}

	if len(dynamic.Env.Image) != 0 {
		images, err := filterExprList(data, dynamic.Env.Image)
		if err != nil {
			return fmt.Errorf("filter images: %w", err)
		}
		if len(images) == 0 {
			buildInput.Empty = true
			logE.Info("all image are ignored, so build isn't run")
			return nil
		}
		dynamic.Env.Image = images
	}

	if len(dynamic.Env.ComputeType) != 0 {
		computeTypes, err := filterExprList(data, dynamic.Env.ComputeType)
		if err != nil {
			return fmt.Errorf("filter compute-type: %w", err)
		}
		if len(computeTypes) == 0 {
			buildInput.Empty = true
			logE.Info("all compute-type are ignored, so build isn't run")
			return nil
		}
		dynamic.Env.ComputeType = computeTypes
	}

	if len(dynamic.Env.Variables) != 0 {
		envVars := make(map[string]bspec.ExprList, len(dynamic.Env.Variables))
		for k, v := range dynamic.Env.Variables {
			vars, err := filterExprList(data, v)
			if err != nil {
				return fmt.Errorf("filter env.variables: %w", err)
			}
			if len(vars) == 0 {
				buildInput.Empty = true
				logE.WithFields(logrus.Fields{
					"env_name": k,
				}).Info("all environment variable are ignored, so build isn't run")
				return nil
			}
			envVars[k] = vars
		}
		dynamic.Env.Variables = envVars
	}

	buildspec.Batch.BuildMatrix.Dynamic = dynamic

	if len(dynamic.Buildspec) > 1 || len(dynamic.Env.Image) > 1 || len(dynamic.Env.ComputeType) > 1 || getSizeOfEnvVars(dynamic.Env.Variables) > 1 {
		// batch build
		buildInput.Batched = true
		if err := setBatchBuildInput(buildInput.BatchBuild, buildspec, data); err != nil {
			return fmt.Errorf("set codebuild.StartBuildBatchInput: %w", err)
		}
		return nil
	}

	// build
	if err := setMatrixBuildInput(data, buildStatusContext, dynamic, buildspec.Lambuild, buildInput.Build); err != nil {
		return fmt.Errorf("set codebuild.StartBuildInput: %w", err)
	}

	return nil
}

func filterExprList(data *domain.Data, src bspec.ExprList) (bspec.ExprList, error) {
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
		f, err := domain.RunExpr(a.If, data)
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

func setMatrixBuildInput(data *domain.Data, buildStatusContext *template.Template, dynamic bspec.MatrixDynamic, lambuild bspec.Lambuild, input *codebuild.StartBuildInput) error {
	if err := setBuildStatusContext(buildStatusContext, data, input); err != nil {
		return err
	}

	if len(dynamic.Buildspec) != 0 {
		input.BuildspecOverride = aws.String(dynamic.Buildspec[0].(string))
	}
	if len(dynamic.Env.Image) != 0 {
		input.ImageOverride = aws.String(dynamic.Env.Image[0].(string))
	}
	if len(dynamic.Env.ComputeType) != 0 {
		input.ComputeTypeOverride = aws.String(dynamic.Env.ComputeType[0].(string))
	}

	envMap := make(map[string]string, len(lambuild.Env.Variables))
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
	if getSizeOfEnvVars(dynamic.Env.Variables) != 0 {
		for k, v := range dynamic.Env.Variables {
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
