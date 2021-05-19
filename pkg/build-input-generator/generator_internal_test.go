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
)

func Test_setBatchBuildInput(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		exp       codebuild.StartBuildBatchInput
		data      domain.Data
		buildspec bspec.Buildspec
	}{
		{
			title: "minimum",
		},
		{
			title: "normal",
			buildspec: bspec.Buildspec{
				Lambuild: bspec.Lambuild{
					Env: bspec.LambuildEnv{
						Variables: map[string]expr.String{
							"FOO": expr.NewStringForTest(t, `"BAR"`),
						},
					},
				},
			},
			exp: codebuild.StartBuildBatchInput{
				EnvironmentVariablesOverride: []*codebuild.EnvironmentVariable{
					{
						Name:  aws.String("FOO"),
						Value: aws.String("BAR"),
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			input := codebuild.StartBuildBatchInput{}
			if err := setBatchBuildInput(&input, d.buildspec, &d.data); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, input, cmpopts.IgnoreFields(codebuild.StartBuildBatchInput{}, "BuildspecOverride")); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
