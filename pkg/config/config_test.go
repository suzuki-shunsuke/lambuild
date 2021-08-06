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
			if hook.If.Empty() {
				t.Fatal("hook.If is nil")
			}
		})
	}
}

func TestConfig_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		src   string
	}{
		{
			title: "normal",
			src: `region: us-east-1
repositories:
  - name: suzuki-shunsuke/lambuild
    hooks:
      - {}
log-level: debug
build-status-context: yoo
error-notification-template: foo
secrets-manager:
  secret-id: lambuild
`,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			cfg := config.Config{}
			if err := yaml.Unmarshal([]byte(d.src), &cfg); err != nil {
				t.Fatal(err)
			}
			if cfg.LogLevel.Get() == 0 {
				t.Fatal("config.LogLevel is empty")
			}
			if cfg.BuildStatusContext.Empty() {
				t.Fatal("cfg.BuildStatusContext is nil")
			}
			if cfg.ErrorNotificationTemplate.Empty() {
				t.Fatal("cfg.ErrorNotificationTemplate is nil")
			}
		})
	}
}
