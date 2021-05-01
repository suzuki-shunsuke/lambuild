package buildspec

type Lambuild struct {
	Filter []LambuildFilter
}

type LambuildFilter struct {
	Event   []string
	File    string
	Label   string
	Ref     string
	BaseRef string `yaml:"base-ref"`
	Author  string
}

type Buildspec struct {
	Version float64 `yaml:"version"`
	Env     Env     `yaml:",omitempty"`
	Phases  Phases  `yaml:",omitempty"`
	Batch   Batch   `yaml:",omitempty"`
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
