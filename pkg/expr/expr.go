package expr

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Expr struct {
	prog *vm.Program
}

func (ex *Expr) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var a string
	if err := unmarshal(&a); err != nil {
		return fmt.Errorf("expression must be a string: %w", err)
	}
	prog, err := expr.Compile(a)
	if err != nil {
		return fmt.Errorf("compile a program: %w", err)
	}
	ex.prog = prog
	return nil
}

func New(s string) (Expr, error) {
	prog, err := expr.Compile(s)
	if err != nil {
		return Expr{}, fmt.Errorf("compile a program: %w", err)
	}
	return Expr{prog: prog}, nil
}

func (ex *Expr) Empty() bool {
	return ex.prog == nil
}

func (ex *Expr) Run(param interface{}) (interface{}, error) {
	a, err := expr.Run(ex.prog, param)
	if err != nil {
		return false, fmt.Errorf("evaluate a expr's compiled program: %w", err)
	}
	return a, nil
}
