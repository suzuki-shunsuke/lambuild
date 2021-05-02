package buildspec

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type ListElement struct {
	Identifier    string
	Buildspec     string      `yaml:",omitempty"`
	Env           ListEnv     `yaml:",omitempty"`
	DebugSession  bool        `yaml:"debug-session,omitempty"`
	IgnoreFailure bool        `yaml:"ignore-failure,omitempty"`
	If            *vm.Program `yaml:",omitempty"`
}

type ListEnv struct {
	ComputeType    string            `yaml:"compute-type,omitempty"`
	Image          string            `yaml:",omitempty"`
	Type           string            `yaml:",omitempty"`
	Variables      map[string]string `yaml:",omitempty"`
	PrivilegedMode bool              `yaml:"privileged-mode,omitempty"`
}

func (list *ListElement) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias ListElement
	a := struct {
		*alias
		If string
	}{
		alias: (*alias)(list),
	}
	if err := unmarshal(&a); err != nil {
		return err
	}
	if a.If != "" {
		prog, err := expr.Compile(a.If, expr.AsBool())
		if err != nil {
			return fmt.Errorf("compile an expression: %s: %w", a.If, err)
		}
		list.If = prog
	}
	return nil
}
