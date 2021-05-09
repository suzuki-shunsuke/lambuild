package buildspec

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Buildspec struct {
	Batch    Batch                  `yaml:",omitempty"`
	Lambuild Lambuild               `yaml:",omitempty"`
	Map      map[string]interface{} `yaml:"-"`
}

func (buildspec *Buildspec) MarshalYAML() (interface{}, error) {
	m := make(map[string]interface{}, len(buildspec.Map))
	for k, v := range buildspec.Map {
		if k == "lambuild" {
			continue
		}
		m[k] = v
	}
	m["batch"] = buildspec.Batch
	return m, nil
}

type Lambuild struct {
	Env LambuildEnv
}

type Batch struct {
	BuildGraph  []GraphElement `yaml:"build-graph,omitempty"`
	BuildList   []ListElement  `yaml:"build-list,omitempty"`
	BuildMatrix Matrix         `yaml:"build-matrix,omitempty"`
}

type Env struct {
	GitCredentialHelper string            `yaml:"git-credential-helper,omitempty"`
	SecretsManager      map[string]string `yaml:"secrets-manager,omitempty"`
	Variables           map[string]string `yaml:",omitempty"`
}

type Phases struct {
	Build Phase `yaml:",omitempty"`
}

type Phase struct {
	Commands []string `yaml:",omiempty"`
}

type LambuildEnv struct {
	Variables map[string]*vm.Program
}

func (env *LambuildEnv) UnmarshalYAML(unmarshal func(interface{}) error) error {
	a := struct {
		Variables map[string]string
	}{}
	if err := unmarshal(&a); err != nil {
		return err
	}
	vars := make(map[string]*vm.Program, len(a.Variables))
	for k, v := range a.Variables {
		prog, err := expr.Compile(v)
		if err != nil {
			return fmt.Errorf("compile an expression: %s: %w", v, err)
		}
		vars[k] = prog
	}
	env.Variables = vars
	return nil
}
