package buildspec_test

import (
	"testing"

	"github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"gopkg.in/yaml.v2"
)

func TestListElement_UnmarshalYAML(t *testing.T) {
	data := []struct {
		title string
		src   string
	}{
		{
			title: "normal",
			src: `if: true
identifier: validate
`,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			elem := buildspec.ListElement{}
			if err := yaml.Unmarshal([]byte(d.src), &elem); err != nil {
				t.Fatal(err)
			}
			if elem.Identifier == "" {
				t.Fatal("elem.Identifier is empty")
			}
			if elem.If == nil {
				t.Fatal("elem.If is nil")
			}
		})
	}
}
