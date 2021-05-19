package generator

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-cmp/cmp"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"github.com/suzuki-shunsuke/lambuild/pkg/mutex"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func Test_handleMatrix(t *testing.T) {
	t.Parallel()
	data := []struct {
		title      string
		data       domain.Data
		expression string
		exp        bool
	}{
		{
			title: "normal",
			data: domain.Data{
				PullRequest: domain.PullRequest{
					ChangedFileNames: mutex.NewStringList("modules/README.md"),
					LabelNames:       mutex.NewStringList(""),
				},
				Event: domain.Event{
					Headers: domain.Headers{
						Event: "pull_request",
					},
				},
			},
			expression: `
			event.Headers.Event == "push" ||
			any(getPRLabelNames(), {# in ["api"]}) ||
			any(getPRFileNames(), {# startsWith "api/"}) ||
			any(getPRFileNames(), {# matches "^modules/"})`,
			exp: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			exp, err := expr.NewBool(d.expression)
			if err != nil {
				t.Fatal(err)
			}

			f, err := exp.Run(d.data.Convert())
			if err != nil {
				t.Fatal(err)
			}
			if (f && !d.exp) || (!f && d.exp) {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}

func Test_getSizeOfEnvVars(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		exp   int
		m     map[string]bspec.ExprList
	}{
		{
			title: "normal",
			exp:   6,
			m: map[string]bspec.ExprList{
				"FOO": {
					"foo1",
					"foo2",
				},
				"BAR": {
					"bar1",
					"bar2",
					"bar3",
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			size := getSizeOfEnvVars(d.m)
			if d.exp != size {
				t.Fatalf("got %d, wanted %d", size, d.exp)
			}
		})
	}
}

func Test_filterExprList(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		data  domain.Data
		src   bspec.ExprList
		exp   bspec.ExprList
	}{
		{
			title: "normal",
			src: bspec.ExprList{
				"foo",
				bspec.ExprElem{
					Value: "bar",
				},
			},
			exp: bspec.ExprList{
				"foo",
				"bar",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			list, err := filterExprList(&d.data, d.src)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(d.exp, list) {
				t.Fatalf("got %+v, wanted %+v", list, d.exp)
			}
		})
	}
}

func Test_setMatrixBuildInput(t *testing.T) {
	t.Parallel()
	data := []struct {
		title              string
		data               domain.Data
		buildStatusContext template.Template
		dynamic            bspec.MatrixDynamic
		lambuild           bspec.Lambuild
		exp                codebuild.StartBuildInput
	}{
		{
			title: "minimum",
		},
		{
			title: "normal",
			dynamic: bspec.MatrixDynamic{
				Buildspec: bspec.ExprList{
					"foo.yml",
				},
				Env: bspec.MatrixDynamicEnv{
					Image: bspec.ExprList{
						"alpine",
					},
					ComputeType: bspec.ExprList{
						"BUILD_GENERAL1_SMALL",
					},
					Variables: map[string]bspec.ExprList{
						"FOO": {
							"YOO",
						},
					},
				},
			},
			exp: codebuild.StartBuildInput{
				BuildspecOverride:   aws.String("foo.yml"),
				ImageOverride:       aws.String("alpine"),
				ComputeTypeOverride: aws.String("BUILD_GENERAL1_SMALL"),
				EnvironmentVariablesOverride: []*codebuild.EnvironmentVariable{
					{
						Name:  aws.String("FOO"),
						Value: aws.String("YOO"),
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			input := codebuild.StartBuildInput{}
			err := setMatrixBuildInput(&d.data, d.buildStatusContext, d.dynamic, d.lambuild, &input)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, input); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
