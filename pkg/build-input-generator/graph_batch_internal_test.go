package generator

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
)

func Test_setGraphBuildInput(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		input     codebuild.StartBuildInput
		buildspec bspec.Buildspec
		data      domain.Data
		elem      bspec.GraphElement
		isErr     bool
		exp       codebuild.StartBuildInput
	}{
		{
			title: "minimum",
			exp: codebuild.StartBuildInput{
				EnvironmentVariablesOverride: []*codebuild.EnvironmentVariable{},
			},
		},
		{
			title: "normal",
			elem: bspec.GraphElement{
				DebugSession: true,
				Env: bspec.GraphEnv{
					Image:          "alpine:3.13.5",
					ComputeType:    "BUILD_GENERAL1_SMALL",
					PrivilegedMode: true,
					Variables: map[string]string{
						"FOO": "FOO_VALUE",
					},
				},
			},
			exp: codebuild.StartBuildInput{
				DebugSessionEnabled:    aws.Bool(true),
				PrivilegedModeOverride: aws.Bool(true),
				ComputeTypeOverride:    aws.String("BUILD_GENERAL1_SMALL"),
				ImageOverride:          aws.String("alpine:3.13.5"),
				EnvironmentVariablesOverride: []*codebuild.EnvironmentVariable{
					{
						Name:  aws.String("FOO"),
						Value: aws.String("FOO_VALUE"),
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := setGraphBuildInput(&d.input, nil, &d.data, d.elem)
			if d.isErr {
				if err == nil {
					t.Fatal("err must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(d.exp, d.input) {
				t.Fatalf("got %+v, wanted %+v", d.input, d.exp)
			}
		})
	}
}
