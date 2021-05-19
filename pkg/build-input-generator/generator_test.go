package generator_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sirupsen/logrus"
	generator "github.com/suzuki-shunsuke/lambuild/pkg/build-input-generator"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func newBool(s string) expr.Bool {
	b, err := expr.NewBool(s)
	panicErr(err)
	return b
}

func TestGenerateInput(t *testing.T) {
	t.Parallel()
	data := []struct {
		title              string
		data               *domain.Data
		buildspec          bspec.Buildspec
		buildStatusContext template.Template
		repo               config.Repository
		exp                domain.BuildInput
	}{
		{
			title:     "minimum",
			data:      &domain.Data{},
			buildspec: bspec.Buildspec{},
			repo:      config.Repository{},
			exp: domain.BuildInput{
				Builds: []*codebuild.StartBuildInput{
					{},
				},
				BatchBuild: &codebuild.StartBuildBatchInput{},
			},
		},
		{
			title: "buildspec is ignored",
			data:  &domain.Data{},
			buildspec: bspec.Buildspec{
				Lambuild: bspec.Lambuild{
					If: newBool("false"),
				},
			},
			repo: config.Repository{},
			exp: domain.BuildInput{
				Empty: true,
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			input, err := generator.GenerateInput(logrus.WithFields(logrus.Fields{}), d.buildStatusContext, d.data, d.buildspec, d.repo)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(input, d.exp, cmpopts.IgnoreFields(codebuild.StartBuildInput{}, "BuildspecOverride")); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
