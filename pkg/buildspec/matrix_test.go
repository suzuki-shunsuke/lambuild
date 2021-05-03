package buildspec_test

import (
	"testing"

	"github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"gopkg.in/yaml.v2"
)

func TestExprList_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		src   string
	}{
		{
			title: "normal",
			src: `- foo
- value: yoo
  if: "true"
`,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			list := buildspec.ExprList{}
			if err := yaml.Unmarshal([]byte(d.src), &list); err != nil {
				t.Fatal(err)
			}
			if len(list) == 0 {
				t.Fatal("list is empty")
			}
		})
	}
}
