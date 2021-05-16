package buildspec_test

import (
	"reflect"
	"testing"

	"github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"gopkg.in/yaml.v2"
)

func TestPhase_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		title            string
		yaml             string
		exp              map[string]interface{}
		numberOfCommands int
		numberOfFinally  int
	}{
		{
			title: "normal",
			yaml: `
commands:
- echo commands
finally:
- echo finally
run-as: root
on-failure: ABORT
runtime-versions:
  java: corretto8
  python: 3.x
  ruby: "$MY_RUBY_VAR"
`,
			exp: map[string]interface{}{
				"run-as":     "root",
				"on-failure": "ABORT",
				"runtime-versions": map[interface{}]interface{}{
					"java":   "corretto8",
					"python": "3.x",
					"ruby":   "$MY_RUBY_VAR",
				},
			},
			numberOfCommands: 1,
			numberOfFinally:  1,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			phase := buildspec.Phase{}
			if err := yaml.Unmarshal([]byte(d.yaml), &phase); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(d.exp, phase.Map) {
				t.Fatalf("got %+v, wanted %+v", phase.Map, d.exp)
			}
			if len(phase.Commands) != d.numberOfCommands {
				t.Fatalf("len(phase.Commands) = %d, wanted %d", len(phase.Commands), d.numberOfCommands)
			}
			if len(phase.Finally) != d.numberOfFinally {
				t.Fatalf("len(phase.Finally) = %d, wanted %d", len(phase.Finally), d.numberOfFinally)
			}
		})
	}
}
