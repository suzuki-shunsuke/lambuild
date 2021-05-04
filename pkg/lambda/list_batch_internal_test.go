package lambda

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
)

func TestHandler_setListBuildInput(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		input     codebuild.StartBuildInput
		buildspec bspec.Buildspec
		data      Data
		elem      bspec.ListElement
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
			elem: bspec.ListElement{
				DebugSession: true,
				Env: bspec.ListEnv{
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
			handler := Handler{}
			err := handler.setListBuildInput(&d.input, &d.data, d.elem)
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
