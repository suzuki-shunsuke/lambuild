package lambda

import (
	"regexp"
	"testing"

	"github.com/antonmedv/expr"
	wh "github.com/suzuki-shunsuke/lambuild/pkg/webhook"
)

func Test_handler_handleMatrix(t *testing.T) {
	data := []struct {
		title      string
		event      wh.Event
		webhook    wh.Webhook
		expression string
		exp        bool
	}{
		{
			title: "normal",
			event: wh.Event{
				// Labels:           []string{"api"},
				ChangedFileNames: []string{"modules/README.md"},
			},
			webhook: wh.Webhook{
				Headers: wh.Headers{
					Event: "pull_request",
				},
			},
			expression: `
			webhook.Headers.Event == "push" ||
			any(event.Labels, {# in ["api"]}) ||
			any(event.ChangedFileNames, {# startsWith "api/"}) ||
			any(event.ChangedFileNames, {regexp.match("^modules/", #)})`,
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

			a, err := expr.Run(exp, map[string]interface{}{
				"event":   d.event,
				"webhook": d.webhook,
				"regexp": map[string]interface{}{
					"match": func(pattern, s string) bool {
						f, err := regexp.MatchString(pattern, s)
						if err != nil {
							panic(err)
						}
						return f
					},
				},
			})
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
