package generator

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
	"gopkg.in/yaml.v2"
)

func Test_handleList(t *testing.T) {
	t.Parallel()
	data := []struct {
		title              string
		input              domain.BuildInput
		buildStatusContext template.Template
		data               domain.Data
		buildspec          bspec.Buildspec
		isErr              bool
		exp                domain.BuildInput
	}{
		{
			title: "minimum",
			exp: domain.BuildInput{
				Empty: true,
			},
		},
	}
	logE := logrus.WithFields(logrus.Fields{})
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := handleList(&d.input, logE, d.buildStatusContext, &d.data, d.buildspec)
			if d.isErr {
				if err == nil {
					t.Fatal("err must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, d.input); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_setListBuildInput(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		input codebuild.StartBuildInput
		data  domain.Data
		elem  bspec.ListElement
		isErr bool
		exp   codebuild.StartBuildInput
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
			err := setListBuildInput(&d.input, template.Template{}, &d.data, bspec.Lambuild{}, d.elem)
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

func Test_extractBuildList(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		input string
		exp   []string
		data  domain.Data
		isErr bool
	}{
		{
			title: "normal",
			input: `
- identifier: test
  if: "false"
- identifier: deploy
  if: "1 == 1"
- identifier: build
`,
			exp: []string{"deploy", "build"},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()

			allElems := []bspec.ListElement{}
			if err := yaml.Unmarshal([]byte(d.input), &allElems); err != nil {
				t.Fatal(err)
			}

			elems, err := extractBuildList(&d.data, allElems)
			if d.isErr {
				if err == nil {
					t.Fatal("err must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			names := make([]string, len(elems))
			for i, elem := range elems {
				names[i] = elem.Identifier
			}
			if !reflect.DeepEqual(names, d.exp) {
				t.Fatalf("got %+v, wanted %+v", names, d.exp)
			}
		})
	}
}

func Test_getLambuildEnvVars(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		input string
		exp   []*codebuild.EnvironmentVariable
		isErr bool
	}{
		{
			title: "normal",
			input: `
variables:
  FOO: '"foo"'
`,
			exp: []*codebuild.EnvironmentVariable{
				{
					Name:  aws.String("FOO"),
					Value: aws.String("foo"),
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()

			env := bspec.LambuildEnv{}
			if err := yaml.Unmarshal([]byte(d.input), &env); err != nil {
				t.Fatal(err)
			}

			data := domain.Data{}

			envs, err := getLambuildEnvVars(&data, bspec.Lambuild{
				Env: env,
			})
			if d.isErr {
				if err == nil {
					t.Fatal("err must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(envs, d.exp) {
				t.Fatalf("got %+v, wanted %+v", envs, d.exp)
			}
		})
	}
}
