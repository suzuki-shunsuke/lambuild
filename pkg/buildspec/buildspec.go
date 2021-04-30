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
	Version float64 `json:"version"`
	Env     Env     `json:"env"`
	Phases  Phases  `json:"phases"`
	Batch   Batch   `json:"batch"`
}

type Batch struct {
	BuildGraph []GraphElement `json:"build-graph" yaml:"build-graph"`
}

type GraphElement struct {
	Identifier    string   `json:"identifier"`
	Buildspec     string   `json:"buildspec"`
	DependOn      []string `json:"depend-on" yaml:"depend-on"`
	Env           GraphEnv `json:"env"`
	DebugSession  bool     `json:"debug-session" yaml:"debug-session"`
	IgnoreFailure bool     `json:"ignore-failure" yaml:"ignore-failure"`
	Lambuild      Lambuild `json:"lambuild,omitempty" yaml:"lambuild,omitempty"`
}

type GraphEnv struct {
	ComputeType    string            `json:"compute-type" yaml:"compute-type"`
	Image          string            `json:"image"`
	Type           string            `json:"type"`
	Variables      map[string]string `json:"variables"`
	PrivilegedMode bool              `json:"privileged-mode" yaml:"privileged-mode"`
}

type Env struct {
	GitCredentialHelper string            `json:"git-credential-helper" yaml:"git-credential-helper"`
	SecretsManager      map[string]string `json:"secrets-manager" yaml:"secrets-manager"`
	Variables           map[string]string `json:"variables"`
}

type Phases struct {
	Build Phase `json:"build"`
}

type Phase struct {
	Commands []string `json:"commands"`
}
