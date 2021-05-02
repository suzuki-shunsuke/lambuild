package buildspec

import (
	"errors"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Matrix struct {
	Static  MatrixStatic
	Dynamic MatrixDynamic
}

type MatrixStatic struct {
	IgnoreFailure bool            `yaml:"ignore-failure,omitempty"`
	Env           MatrixStaticEnv `yaml:",omitempty"`
}

type MatrixDynamic struct {
	Buildspec ExprList         `yaml:",omitempty"`
	Env       MatrixDynamicEnv `yaml:",omitempty"`
}

type ExprList []interface{}

type ExprElem struct {
	Value string
	If    *vm.Program
}

func (list *ExprList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val []interface{}
	if err := unmarshal(&val); err != nil {
		return err
	}
	arr := make(ExprList, len(val))
	for i, a := range val {
		if _, ok := a.(string); ok {
			arr[i] = a
			continue
		}
		b, ok := a.(map[interface{}]interface{})
		if !ok {
			return errors.New("buildspec must be string or map[interface{}]interface{}")
		}
		bp := ExprElem{}
		for k, v := range b {
			ks, ok := k.(string)
			if !ok {
				return errors.New("key of buildspec must be string")
			}
			switch ks {
			case "value":
				vs, ok := v.(string)
				if !ok {
					return errors.New("value of buildspec must be string")
				}
				bp.Value = vs
			case "if":
				vs, ok := v.(string)
				if !ok {
					return errors.New("if of buildspec must be string")
				}
				prog, err := expr.Compile(vs, expr.AsBool())
				if err != nil {
					return fmt.Errorf("compile an expression: %s: %w", vs, err)
				}
				bp.If = prog
			default:
				return errors.New("invalid key of buildspec: " + ks)
			}
		}
		arr[i] = bp
	}
	*list = arr
	return nil
}

type MatrixDynamicEnv struct {
	ComputeType ExprList            `yaml:"compute-type,omitempty"`
	Image       ExprList            `yaml:",omitempty"`
	Variables   map[string]ExprList `yaml:",omitempty"`
}

type MatrixStaticEnv struct {
	ComputeType    string            `yaml:"compute-type,omitempty"`
	Image          string            `yaml:",omitempty"`
	Type           string            `yaml:",omitempty"`
	Variables      map[string]string `yaml:",omitempty"`
	PrivilegedMode bool              `yaml:"privileged-mode,omitempty"`
}

func (matrix *Matrix) Empty() bool {
	if len(matrix.Dynamic.Buildspec) != 0 {
		return false
	}
	if len(matrix.Dynamic.Env.ComputeType) != 0 {
		return false
	}
	if len(matrix.Dynamic.Env.Image) != 0 {
		return false
	}
	if len(matrix.Dynamic.Env.Variables) != 0 {
		return false
	}
	return true
}
