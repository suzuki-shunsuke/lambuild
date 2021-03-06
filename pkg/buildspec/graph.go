package buildspec

import (
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
)

type GraphElement struct {
	Identifier    string
	Buildspec     string    `yaml:",omitempty"`
	DependOn      []string  `yaml:"depend-on,omitempty"`
	Env           GraphEnv  `yaml:",omitempty"`
	DebugSession  bool      `yaml:"debug-session,omitempty"`
	IgnoreFailure bool      `yaml:"ignore-failure,omitempty"`
	If            expr.Bool `yaml:",omitempty"`
}

type GraphEnv struct {
	ComputeType    string            `yaml:"compute-type,omitempty"`
	Image          string            `yaml:",omitempty"`
	Type           string            `yaml:",omitempty"`
	Variables      map[string]string `yaml:",omitempty"`
	PrivilegedMode bool              `yaml:"privileged-mode,omitempty"`
}
