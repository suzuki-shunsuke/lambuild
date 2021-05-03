package lambda

import (
	"testing"

	"github.com/antonmedv/expr"
)

func Test_handler_handleMatrix(t *testing.T) {
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
			any(getPRFileNames(), {regexp.match("^modules/", #)})`,
			exp: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			exp, err := expr.Compile(d.expression, expr.AsBool())
			if err != nil {
				t.Fatal(err)
			}

			a, err := runExpr(exp, &d.data)
			if err != nil {
				t.Fatal(err)
			}
			f := a.(bool)
			if (f && !d.exp) || (!f && d.exp) {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
