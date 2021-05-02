package config

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Config struct {
	Repositories []Repository
}

type Repository struct {
	Name      string
	Hooks     []Hook
	CodeBuild CodeBuild `yaml:"codebuild"`
}

type CodeBuild struct {
	ProjectName string `yaml:"project-name"`
}

type Hook struct {
	If     *vm.Program
	Config string
}

func (hook *Hook) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Hook
	a := struct {
		*alias
		If string
	}{
		alias: (*alias)(hook),
	}
	if err := unmarshal(&a); err != nil {
		return err
	}
	if a.If != "" {
		prog, err := expr.Compile(a.If, expr.AsBool())
		if err != nil {
			return fmt.Errorf("compile an expression: %s: %w", a.If, err)
		}
		hook.If = prog
	}
	return nil
}
