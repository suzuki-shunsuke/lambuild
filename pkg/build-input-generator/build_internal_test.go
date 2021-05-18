package generator

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func newBool(s string) expr.Bool {
	b, err := expr.NewBool(s)
	if err != nil {
		panic(err)
	}
	return b
}

func newTemplate(s string) template.Template {
	b, err := template.New(s)
	if err != nil {
		panic(err)
	}
	return b
}

func Test_handleBuildItem(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		data      *domain.Data
		buildspec bspec.Buildspec
		item      bspec.Item
		exp       codebuild.StartBuildInput
	}{
		{
			title:     "minimum",
			data:      &domain.Data{},
			buildspec: bspec.Buildspec{},
			item:      bspec.Item{},
			exp:       codebuild.StartBuildInput{},
		},
		{
			title:     "item.if is false",
			data:      &domain.Data{},
			buildspec: bspec.Buildspec{},
			item: bspec.Item{
				If:    newBool("false"),
				Image: "alpine", // ignored
			},
			exp: codebuild.StartBuildInput{},
		},
		{
			title: "normal",
			data:  &domain.Data{},
			buildspec: bspec.Buildspec{
				Lambuild: bspec.Lambuild{
					DebugSession:   aws.Bool(true),
					PrivilegedMode: aws.Bool(true),
					GitCloneDepth:  aws.Int64(5),
				},
			},
			item: bspec.Item{
				If:                 newBool("true"),
				Image:              "alpine",
				ComputeType:        "BUILD_GENERAL1_SMALL",
				EnvironmentType:    "LINUX_CONTAINER",
				BuildStatusContext: newTemplate("foo"),
			},
			exp: codebuild.StartBuildInput{
				BuildStatusConfigOverride: &codebuild.BuildStatusConfig{
					Context: aws.String("foo"),
				},
				ImageOverride:           aws.String("alpine"),
				ComputeTypeOverride:     aws.String("BUILD_GENERAL1_SMALL"),
				EnvironmentTypeOverride: aws.String("LINUX_CONTAINER"),
				DebugSessionEnabled:     aws.Bool(true),
				PrivilegedModeOverride:  aws.Bool(true),
				GitCloneDepthOverride:   aws.Int64(5),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			input, err := handleBuildItem(d.data, d.buildspec, d.item)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(input, d.exp, cmpopts.IgnoreFields(codebuild.StartBuildInput{}, "BuildspecOverride")); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
