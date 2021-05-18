package expr

import (
	"errors"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Bool struct {
	prog *vm.Program
}

func (boolExpr *Bool) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var a string
	if err := unmarshal(&a); err != nil {
		return fmt.Errorf("expression must be a string: %w", err)
	}
	prog, err := expr.Compile(a, expr.AsBool())
	if err != nil {
		return fmt.Errorf("compile a program: %w", err)
	}
	boolExpr.prog = prog
	return nil
}

func NewBool(s string) (Bool, error) {
	prog, err := expr.Compile(s, expr.AsBool())
	if err != nil {
		return Bool{}, fmt.Errorf("compile a program: %w", err)
	}
	return Bool{prog: prog}, nil
}

func (boolExpr *Bool) Empty() bool {
	return boolExpr.prog == nil
}

func (boolExpr *Bool) Run(param interface{}) (bool, error) {
	a, err := expr.Run(boolExpr.prog, param)
	if err != nil {
		return false, fmt.Errorf("evaluate a expr's compiled program: %w", err)
	}
	f, ok := a.(bool)
	if !ok {
		return false, errors.New("evaluated result must be bool")
	}
	return f, nil
}
