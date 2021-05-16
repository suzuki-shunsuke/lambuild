package buildspec

import (
	"fmt"

	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
	"gopkg.in/yaml.v2"
)

type Buildspec struct {
	Batch    Batch                  `yaml:",omitempty"`
	Lambuild Lambuild               `yaml:",omitempty"`
	Map      map[string]interface{} `yaml:"-"`
	Phases   Phases
}

func (buildspec *Buildspec) filter(param interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{}, len(buildspec.Map)+2) //nolint:gomnd
	for k, v := range buildspec.Map {
		if k == "lambuild" {
			continue
		}
		m[k] = v
	}
	m["batch"] = buildspec.Batch
	phases, err := buildspec.Phases.Filter(param)
	if err != nil {
		return nil, err
	}
	m["phases"] = phases
	return m, nil
}

func (buildspec *Buildspec) ToYAML(param interface{}) ([]byte, error) {
	m, err := buildspec.filter(param)
	if err != nil {
		return nil, fmt.Errorf("filter commands from buildspec: %w", err)
	}

	builtContent, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal a buildspec: %w", err)
	}
	return builtContent, nil
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
	Env                LambuildEnv
	BuildStatusContext template.Template `yaml:"build-status-context"`
	Image              string
	ComputeType        string `yaml:"compute-type"`
	EnvironmentType    string `yaml:"environment-type"`
	DebugSession       *bool  `yaml:"debug-session"`
	PrivilegedMode     *bool  `yaml:"privileged-mode"`
	GitCloneDepth      *int64 `yaml:"git-clone-depth"`
	ReportBuildStatus  *bool  `yaml:"report-build-status"`
	// It is danger to allow to override Service Role
	// So lambuild doesn't support to override Service Role
	Items []Item
	If    expr.Bool
}

type Item struct {
	If                 expr.Bool
	Env                LambuildEnv
	BuildStatusContext template.Template `yaml:"build-status-context"`
	Image              string
	ComputeType        string `yaml:"compute-type"`
	EnvironmentType    string `yaml:"environment-type"`
	Param              map[string]interface{}
}

type LambuildEnv struct {
	Variables map[string]expr.String
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
	Install   Phase `yaml:",omitempty"`
	PreBuild  Phase `yaml:"pre_build,omitempty"`
	Build     Phase `yaml:",omitempty"`
	PostBuild Phase `yaml:"post_build,omitempty"`
}
