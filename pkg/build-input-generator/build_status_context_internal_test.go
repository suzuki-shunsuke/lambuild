package generator

import (
	"reflect"
	"testing"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	templ "github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func Test_getBuildStatusContext(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		tpl   string
		exp   string
		data  domain.Data
	}{
		{
			title: "minimum",
		},
		{
			title: "normal",
			tpl:   "{{.sha}}",
			exp:   "0000",
			data: domain.Data{
				SHA: "0000",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			var tpl *template.Template
			if d.tpl != "" {
				tp, err := templ.Compile(d.tpl)
				if err != nil {
					t.Fatal(err)
				}
				tpl = tp
			}
			s, err := getBuildStatusContext(tpl, &d.data)
			if err != nil {
				t.Fatal(err)
			}
			if s != d.exp {
				t.Fatalf("got %s, wanted %s", s, d.exp)
			}
		})
	}
}

func Test_setBuildStatusContext(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		tpl   string
		input codebuild.StartBuildInput
		exp   codebuild.StartBuildInput
		data  domain.Data
	}{
		{
			title: "minimum",
		},
		{
			title: "normal",
			tpl:   "{{.sha}}",
			exp: codebuild.StartBuildInput{
				BuildStatusConfigOverride: &codebuild.BuildStatusConfig{
					Context: aws.String("0000"),
				},
			},
			data: domain.Data{
				SHA: "0000",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			var tpl *template.Template
			if d.tpl != "" {
				tl, err := templ.Compile(d.tpl)
				if err != nil {
					t.Fatal(err)
				}
				tpl = tl
			}

			if err := setBuildStatusContext(tpl, &d.data, &d.input); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(d.exp, d.input) {
				t.Fatalf("got %+v, wanted %+v", d.input, d.exp)
			}
		})
	}
}
