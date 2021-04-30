package buildspec

type Lambuild struct {
	Filter []LambuildFilter
}

type LambuildFilter struct {
	Event  []string
	File   string
	Label  string
	Ref    string
	Author string
}

type Buildspec struct {
	Version float64 `yaml:"version"`
	Env     Env     `yaml:",omitempty"`
	Phases  Phases  `yaml:",omitempty"`
	Batch   Batch   `yaml:",omitempty"`
}

type Batch struct {
	BuildGraph []GraphElement `yaml:"build-graph,omitempty"`
}

type GraphElement struct {
	Identifier    string
	Buildspec     string   `yaml:",omitempty"`
	DependOn      []string `yaml:"depend-on,omitempty"`
	Env           GraphEnv `yaml:",omitempty"`
	DebugSession  bool     `yaml:"debug-session,omitempty"`
	IgnoreFailure bool     `yaml:"ignore-failure,omitempty"`
	Lambuild      Lambuild `yaml:"lambuild,omitempty"`
}

type GraphEnv struct {
	ComputeType    string            `yaml:"compute-type,omitempty"`
	Image          string            `yaml:",omitempty"`
	Type           string            `yaml:",omitempty"`
	Variables      map[string]string `yaml:",omitempty"`
	PrivilegedMode bool              `yaml:"privileged-mode,omitempty"`
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
