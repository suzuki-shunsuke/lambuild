package buildspec_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"gopkg.in/yaml.v2"
)

func TestCommand_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		yaml  string
		exp   bspec.Command
	}{
		{
			title: "minimum",
			yaml:  "foo",
			exp: bspec.Command{
				Command: "foo",
			},
		},
		{
			title: "normal",
			yaml: `command: foo
if: "true"`,
			exp: bspec.Command{
				Command: "foo",
				If:      expr.NewBoolForTest(t, "true"),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			cmd := bspec.Command{}
			if err := yaml.Unmarshal([]byte(d.yaml), &cmd); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, cmd, cmpopts.IgnoreUnexported(expr.Bool{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestCommands_Filter(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		commands bspec.Commands
		param    interface{}
		exp      []string
	}{
		{
			title:    "minimum",
			commands: bspec.Commands{},
			exp:      []string{},
		},
		{
			title: "normal",
			commands: bspec.Commands{
				bspec.Command{
					Command: "echo hello",
				},
				bspec.Command{
					Command: "echo false",
					If:      expr.NewBoolForTest(t, "false"),
				},
				bspec.Command{
					Command: "echo true",
					If:      expr.NewBoolForTest(t, "true"),
				},
			},
			exp: []string{
				"echo hello",
				"echo true",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			m, err := d.commands.Filter(d.param)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, m); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
