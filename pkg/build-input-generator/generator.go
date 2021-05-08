package generator

import (
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
)

func GenerateInput(logE *logrus.Entry, buildStatusContext *template.Template, data *domain.Data, buildspec bspec.Buildspec, repo config.Repository) (domain.BuildInput, error) {
	buildInput := domain.BuildInput{
		Build: &codebuild.StartBuildInput{
			ProjectName:   aws.String(repo.CodeBuild.ProjectName),
			SourceVersion: aws.String(data.SHA),
		},
		BatchBuild: &codebuild.StartBuildBatchInput{
			ProjectName:   aws.String(repo.CodeBuild.ProjectName),
			SourceVersion: aws.String(data.SHA),
		},
	}

	if len(buildspec.Batch.BuildGraph) != 0 {
		logE.Debug("handling build-graph")
		if err := handleGraph(buildStatusContext, &buildInput, logE, data, buildspec); err != nil {
			return buildInput, err
		}
	}

	if len(buildspec.Batch.BuildList) != 0 {
		logE.Debug("handling build-list")
		if err := handleList(&buildInput, logE, buildStatusContext, data, buildspec); err != nil {
			return buildInput, err
		}
	}

	if !buildspec.Batch.BuildMatrix.Empty() {
		logE.Debug("handling build-matrix")
		if err := handleMatrix(&buildInput, logE, buildStatusContext, data, buildspec); err != nil {
			return buildInput, err
		}
	}
	return buildInput, nil
}
