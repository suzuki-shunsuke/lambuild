package buildspec

import (
	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
)

type ListElement struct {
	Identifier    string
	Buildspec     string    `yaml:",omitempty"`
	Env           ListEnv   `yaml:",omitempty"`
	DebugSession  bool      `yaml:"debug-session,omitempty"`
	IgnoreFailure bool      `yaml:"ignore-failure,omitempty"`
	If            expr.Bool `yaml:",omitempty"`
}

type ListEnv struct {
	ComputeType    string            `yaml:"compute-type,omitempty"`
	Image          string            `yaml:",omitempty"`
	Type           string            `yaml:",omitempty"`
	Variables      map[string]string `yaml:",omitempty"`
	PrivilegedMode bool              `yaml:"privileged-mode,omitempty"`
}
