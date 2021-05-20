package generator

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
		{
			title: "single build",
			buildspec: bspec.Buildspec{
				Batch: bspec.Batch{
					BuildList: []bspec.ListElement{
						{
							Identifier: "foo",
						},
					},
				},
			},
			exp: domain.BuildInput{
				Builds: []*codebuild.StartBuildInput{
					{},
				},
			},
		},
		{
			title: "batch build",
			input: domain.BuildInput{
				BatchBuild: &codebuild.StartBuildBatchInput{},
			},
			buildspec: bspec.Buildspec{
				Batch: bspec.Batch{
					BuildList: []bspec.ListElement{
						{
							Identifier: "foo",
						},
						{
							Identifier: "bar",
						},
					},
				},
			},
			exp: domain.BuildInput{
				Batched:    true,
				BatchBuild: &codebuild.StartBuildBatchInput{},
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
			if diff := cmp.Diff(d.exp, d.input, cmpopts.IgnoreFields(codebuild.StartBuildInput{}, "BuildspecOverride"), cmpopts.IgnoreFields(codebuild.StartBuildBatchInput{}, "BuildspecOverride")); diff != "" {
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
			exp:   codebuild.StartBuildInput{},
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
			if diff := cmp.Diff(d.exp, d.input); diff != "" {
				t.Fatalf(diff)
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
			if diff := cmp.Diff(names, d.exp); diff != "" {
				t.Fatalf(diff)
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
			if diff := cmp.Diff(envs, d.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
