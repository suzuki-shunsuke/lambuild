package lambda

import (
	"testing"
	"text/template"
)

func Test_getBuildStatusContext(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		tpl   string
		exp   string
		data  Data
	}{
		{
			title: "minimum",
		},
		{
			title: "normal",
			tpl:   "{{.sha}}",
			exp:   "0000",
			data: Data{
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
				tp, err := template.New("_").Parse(d.tpl)
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
