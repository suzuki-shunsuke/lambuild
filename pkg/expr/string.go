package expr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type String struct {
	prog *vm.Program
}

func (str *String) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var a string
	if err := unmarshal(&a); err != nil {
		return fmt.Errorf("expression must be a string: %w", err)
	}
	prog, err := expr.Compile(a)
	if err != nil {
		return fmt.Errorf("compile a program: %w", err)
	}
	str.prog = prog
	return nil
}

func NewString(s string) (String, error) {
	prog, err := expr.Compile(s)
	if err != nil {
		return String{}, fmt.Errorf("compile a program: %w", err)
	}
	return String{prog: prog}, nil
}

func NewStringForTest(t *testing.T, s string) String {
	t.Helper()
	a, err := NewString(s)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func (str *String) Empty() bool {
	return str.prog == nil
}

func (str *String) Run(param interface{}) (string, error) {
	a, err := expr.Run(str.prog, param)
	if err != nil {
		return "", fmt.Errorf("evaluate a expr's compiled program: %w", err)
	}
	f, ok := a.(string)
	if !ok {
		return "", errors.New("evaluated result must be string")
	}
	return f, nil
}
