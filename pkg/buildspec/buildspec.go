package buildspec

import (
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
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
	Variables map[string]expr.String
}
