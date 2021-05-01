package lambda

import (
	"context"
	"fmt"

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
	for _, graphElem := range buildspec.Batch.BuildGraph {
		for _, filt := range graphElem.Lambuild.Filter {
			f, err := handler.filter(filt, webhook, event)
			if err != nil {
				return err
			}
			if f {
				graphElems = append(graphElems, graphElem)
				break
			}
		}
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
	} else {
		builtContent, err := yaml.Marshal(handler.generateBuildspec(buildspec, graphElems))
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
	}
	return nil
}
