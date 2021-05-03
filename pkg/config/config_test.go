package config_test

import (
	"testing"

	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"gopkg.in/yaml.v2"
)

func TestHook_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		src   string
	}{
		{
			title: "normal",
			src: `if: true
config: foo.yaml
`,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			hook := config.Hook{}
			if err := yaml.Unmarshal([]byte(d.src), &hook); err != nil {
				t.Fatal(err)
			}
			if hook.Config == "" {
				t.Fatal("hook.Config is empty")
			}
			if hook.If == nil {
				t.Fatal("hook.If is nil")
			}
		})
	}
}
