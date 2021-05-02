package lambda

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	wh "github.com/suzuki-shunsuke/lambuild/pkg/webhook"
	"gopkg.in/yaml.v2"
)

func (handler *Handler) handleGraph(ctx context.Context, logE *logrus.Entry, event wh.Event, webhook wh.Webhook, buildspec bspec.Buildspec, repo config.Repository, envs []*codebuild.EnvironmentVariable) error {
	graphElems := []bspec.GraphElement{}
	for _, elem := range buildspec.Batch.BuildGraph {
		if elem.If == nil {
			graphElems = append(graphElems, elem)
			continue
		}
		f, err := expr.Run(elem.If, setExprFuncs(map[string]interface{}{
			"event":   event,
			"webhook": webhook,
		}))
		if err != nil {
			return fmt.Errorf("evaluate an expression: %w", err)
		}
		if !f.(bool) {
			continue
		}
		elem.If = nil
		graphElems = append(graphElems, elem)
	}
	if len(graphElems) == 0 {
		logE.Info("no graph element is run")
		return nil
	}

	if len(graphElems) == 1 {
		graphElem := graphElems[0]
		input := &codebuild.StartBuildInput{
			BuildspecOverride: aws.String(graphElem.Buildspec),
			ProjectName:       aws.String(repo.CodeBuild.ProjectName),
			SourceVersion:     aws.String(event.SHA),
		}
		if graphElem.DebugSession {
			input.DebugSessionEnabled = aws.Bool(true)
		}

		if err := handler.setBuildStatusContext(event, webhook, input); err != nil {
			return err
		}

		for k, v := range graphElem.Env.Variables {
			envs = append(envs, &codebuild.EnvironmentVariable{
				Name:  aws.String(k),
				Value: aws.String(v),
			})
		}
		input.EnvironmentVariablesOverride = envs

		if graphElem.Env.ComputeType != "" {
			input.ComputeTypeOverride = aws.String(graphElem.Env.ComputeType)
		}

		if graphElem.Env.Image != "" {
			input.ImageOverride = aws.String(graphElem.Env.Image)
		}

		if graphElem.Env.PrivilegedMode {
			input.PrivilegedModeOverride = aws.Bool(true)
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
	buildspec.Batch.BuildGraph = graphElems
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
