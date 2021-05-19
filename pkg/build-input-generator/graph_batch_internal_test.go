package generator

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func Test_handleGraph(t *testing.T) {
	t.Parallel()
	data := []struct {
		title              string
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
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			input := domain.BuildInput{}
			err := handleGraph(d.buildStatusContext, &input, logE, &d.data, d.buildspec)
			if d.isErr {
				if err == nil {
					t.Fatal("err must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, input); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

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
			err := setGraphBuildInput(&d.input, template.Template{}, &d.data, bspec.Lambuild{}, d.elem)
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

func Test_extractGraphByDependency(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		identifiers map[string]bspec.GraphElement
		exp         map[string]bspec.GraphElement
	}{
		{
			title: "minimum",
		},
		{
			title: "remove deploy because test isn't found",
			identifiers: map[string]bspec.GraphElement{
				"build": {},
				"deploy": {
					DependOn: []string{"test"},
				},
			},
			exp: map[string]bspec.GraphElement{
				"build": {},
			},
		},
	}
	logE := logrus.WithFields(logrus.Fields{})
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			extractGraphByDependency(d.identifiers, logE)
			if diff := cmp.Diff(d.exp, d.identifiers, cmpopts.IgnoreUnexported(expr.Bool{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_extractGraphByIf(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		data     *domain.Data
		allElems []bspec.GraphElement
		exp      map[string]bspec.GraphElement
	}{
		{
			title: "minimum",
			exp:   map[string]bspec.GraphElement{},
		},
		{
			title: "normal",
			data:  &domain.Data{},
			allElems: []bspec.GraphElement{
				{
					Identifier: "always",
				},
				{
					Identifier: "false",
					If:         expr.NewBoolForTest(t, "false"),
				},
				{
					Identifier: "true",
					If:         expr.NewBoolForTest(t, "true"),
				},
			},
			exp: map[string]bspec.GraphElement{
				"always": {
					Identifier: "always",
				},
				"true": {
					Identifier: "true",
					If:         expr.NewBoolForTest(t, "true"),
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			identifiers := map[string]bspec.GraphElement{}
			err := extractGraphByIf(d.data, d.allElems, identifiers)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, identifiers, cmpopts.IgnoreUnexported(expr.Bool{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
