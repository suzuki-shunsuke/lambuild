package lambda

import (
	"reflect"
	"testing"

	"github.com/antonmedv/expr"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
)

func Test_handler_handleMatrix(t *testing.T) {
	t.Parallel()
	data := []struct {
		title      string
		data       Data
		expression string
		exp        bool
	}{
		{
			title: "normal",
			data: Data{
				PullRequest: PullRequest{
					ChangedFileNames: []string{"modules/README.md"},
					LabelNames:       []string{},
				},
				Event: Event{
					Headers: Headers{
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
			exp, err := expr.Compile(d.expression, expr.AsBool())
			if err != nil {
				t.Fatal(err)
			}

			a, err := runExpr(exp, &d.data)
			if err != nil {
				t.Fatal(err)
			}
			f := a.(bool) //nolint:forcetypeassert
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
		data  Data
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
