package buildspec_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
)

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
