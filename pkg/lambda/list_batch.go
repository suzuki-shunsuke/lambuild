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

func (handler *Handler) handleList(ctx context.Context, logE *logrus.Entry, event wh.Event, webhook wh.Webhook, buildspec bspec.Buildspec, repo config.Repository, envs []*codebuild.EnvironmentVariable) error {
	listElems := []bspec.ListElement{}
	for _, listElem := range buildspec.Batch.BuildList {
		for _, filt := range listElem.Lambuild.Filter {
			f, err := handler.filter(filt, webhook, event)
			if err != nil {
				return err
			}
			if f {
				listElems = append(listElems, listElem)
				break
			}
		}
	}
	if len(listElems) == 0 {
		logE.Info("no list element is run")
		return nil
	}

	if len(listElems) == 1 {
		listElem := listElems[0]
		input := &codebuild.StartBuildInput{
			BuildspecOverride: aws.String(listElem.Buildspec),
			ProjectName:       aws.String(repo.CodeBuild.ProjectName),
			SourceVersion:     aws.String(event.SHA),
		}
		if listElem.DebugSession {
			input.DebugSessionEnabled = aws.Bool(true)
		}

		for k, v := range listElem.Env.Variables {
			envs = append(envs, &codebuild.EnvironmentVariable{
				Name:  aws.String(k),
				Value: aws.String(v),
			})
		}
		input.EnvironmentVariablesOverride = envs

		if listElem.Env.ComputeType != "" {
			input.ComputeTypeOverride = aws.String(listElem.Env.ComputeType)
		}

		if listElem.Env.Image != "" {
			input.ImageOverride = aws.String(listElem.Env.Image)
		}

		if listElem.Env.PrivilegedMode {
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
		builtContent, err := yaml.Marshal(handler.generateListBuildspec(buildspec, listElems))
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
