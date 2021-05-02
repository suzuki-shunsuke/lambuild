package buildspec

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type GraphElement struct {
	Identifier    string
	Buildspec     string      `yaml:",omitempty"`
	DependOn      []string    `yaml:"depend-on,omitempty"`
	Env           GraphEnv    `yaml:",omitempty"`
	DebugSession  bool        `yaml:"debug-session,omitempty"`
	IgnoreFailure bool        `yaml:"ignore-failure,omitempty"`
	If            *vm.Program `yaml:",omitempty"`
}

type GraphEnv struct {
	ComputeType    string            `yaml:"compute-type,omitempty"`
	Image          string            `yaml:",omitempty"`
	Type           string            `yaml:",omitempty"`
	Variables      map[string]string `yaml:",omitempty"`
	PrivilegedMode bool              `yaml:"privileged-mode,omitempty"`
}

func (graph *GraphElement) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias GraphElement
	a := struct {
		*alias
		If string
	}{
		alias: (*alias)(graph),
	}
	if err := unmarshal(&a); err != nil {
		return err
	}
	if a.If != "" {
		prog, err := expr.Compile(a.If, expr.AsBool())
		if err != nil {
			return fmt.Errorf("compile an expression: %s: %w", a.If, err)
		}
		graph.If = prog
	}
	return nil
}
