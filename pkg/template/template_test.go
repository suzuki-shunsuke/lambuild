package template_test

import (
	"testing"

	"github.com/suzuki-shunsuke/lambuild/pkg/template"
	"gopkg.in/yaml.v2"
)

func TestString_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		yaml  string
		param interface{}
		exp   string
	}{
		{
			title: "normal",
			yaml:  `"{{.name}}"`,
			param: map[string]interface{}{
				"name": "foo",
			},
			exp: "foo",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			b := template.Template{}
			if err := yaml.Unmarshal([]byte(d.yaml), &b); err != nil {
				t.Fatal(err)
			}
			if b.Empty() {
				t.Fatal("string is empty")
			}
			f, err := b.Execute(d.param)
			if err != nil {
				t.Fatal(err)
			}
			if f != d.exp {
				t.Fatalf(`got "%s", wanted "%s"`, f, d.exp)
			}
		})
	}
}

func TestString_Empty(t *testing.T) {
	t.Parallel()
	b := template.Template{}
	if !b.Empty() {
		t.Fatal("Tempalte.Empty() should be true")
	}
}

func TestNewString(t *testing.T) {
	t.Parallel()
	b, err := template.New("foo")
	if err != nil {
		t.Fatal(err)
	}
	f, err := b.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if f != "foo" {
		t.Fatal(`Template must be "foo"`)
	}
}
