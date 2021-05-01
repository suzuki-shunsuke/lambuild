package buildspec

type ListElement struct {
	Identifier    string
	Buildspec     string   `yaml:",omitempty"`
	Env           ListEnv  `yaml:",omitempty"`
	Lambuild      Lambuild `yaml:"lambuild,omitempty"`
	DebugSession  bool     `yaml:"debug-session,omitempty"`
	IgnoreFailure bool     `yaml:"ignore-failure,omitempty"`
}

type ListEnv struct {
	ComputeType    string            `yaml:"compute-type,omitempty"`
	Image          string            `yaml:",omitempty"`
	Type           string            `yaml:",omitempty"`
	Variables      map[string]string `yaml:",omitempty"`
	PrivilegedMode bool              `yaml:"privileged-mode,omitempty"`
}
