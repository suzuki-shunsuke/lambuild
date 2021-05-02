package lambda

import (
	"context"
	"fmt"
	"regexp"

	"github.com/antonmedv/expr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	wh "github.com/suzuki-shunsuke/lambuild/pkg/webhook"
	"gopkg.in/yaml.v2"
)

func setExprFuncs(env map[string]interface{}) map[string]interface{} {
	env["regexp"] = map[string]interface{}{
		"match": func(pattern, s string) bool {
			f, err := regexp.MatchString(pattern, s)
			if err != nil {
				panic(err)
			}
			return f
		},
	}
	return env
}

func (handler *Handler) filterExprList(event wh.Event, webhook wh.Webhook, src bspec.ExprList) (bspec.ExprList, error) {
	list := bspec.ExprList{}
	for _, bs := range src {
		s, ok := bs.(string)
		if ok {
			list = append(list, s)
			continue
		}
		a := bs.(bspec.ExprElem)
		f, err := expr.Run(a.If, setExprFuncs(map[string]interface{}{
			"event":   event,
			"webhook": webhook,
		}))
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

func (handler *Handler) handleMatrix(ctx context.Context, logE *logrus.Entry, event wh.Event, webhook wh.Webhook, buildspec bspec.Buildspec, repo config.Repository, envs []*codebuild.EnvironmentVariable) error {
	buildspecs, err := handler.filterExprList(event, webhook, buildspec.Batch.BuildMatrix.Dynamic.Buildspec)
	if err != nil {
		return fmt.Errorf("filter buildspecs: %w", err)
	}
	buildspec.Batch.BuildMatrix.Dynamic.Buildspec = buildspecs

	images, err := handler.filterExprList(event, webhook, buildspec.Batch.BuildMatrix.Dynamic.Env.Image)
	if err != nil {
		return fmt.Errorf("filter images: %w", err)
	}
	buildspec.Batch.BuildMatrix.Dynamic.Env.Image = images

	computeTypes, err := handler.filterExprList(event, webhook, buildspec.Batch.BuildMatrix.Dynamic.Env.ComputeType)
	if err != nil {
		return fmt.Errorf("filter compute-type: %w", err)
	}
	buildspec.Batch.BuildMatrix.Dynamic.Env.ComputeType = computeTypes

	envVars := make(map[string]bspec.ExprList, len(buildspec.Batch.BuildMatrix.Dynamic.Env.Variables))
	for k, v := range buildspec.Batch.BuildMatrix.Dynamic.Env.Variables {
		vars, err := handler.filterExprList(event, webhook, v)
		if err != nil {
			return fmt.Errorf("filter env.variables: %w", err)
		}
		envVars[k] = vars
	}
	buildspec.Batch.BuildMatrix.Dynamic.Env.Variables = envVars

	if len(buildspecs) > 1 || len(images) > 1 || len(computeTypes) > 1 || getSizeOfEnvVars(envVars) > 1 {
		// batch build
		builtContent, err := yaml.Marshal(buildspec)
		if err != nil {
			return err
		}
		buildOut, err := handler.CodeBuild.StartBuildBatchWithContext(ctx, &codebuild.StartBuildBatchInput{
			BuildspecOverride:            aws.String(string(builtContent)),
			ProjectName:                  aws.String(repo.CodeBuild.ProjectName),
			SourceVersion:                aws.String(event.SHA),
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

	// build
	input := &codebuild.StartBuildInput{
		ProjectName:   aws.String(repo.CodeBuild.ProjectName),
		SourceVersion: aws.String(event.SHA),
	}

	if err := handler.setBuildStatusContext(event, webhook, input); err != nil {
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
	if getSizeOfEnvVars(envVars) != 0 {
		list := make([]*codebuild.EnvironmentVariable, 0, len(envVars))
		for k, v := range envVars {
			list = append(list, &codebuild.EnvironmentVariable{
				Name:  aws.String(k),
				Value: aws.String(v[0].(string)),
			})
		}
		input.EnvironmentVariablesOverride = list
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
