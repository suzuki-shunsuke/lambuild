package config

type Config struct {
	Repositories []Repository
}

type Repository struct {
	Name      string
	Hooks     []Hook
	CodeBuild CodeBuild `yaml:"codebuild"`
}

type CodeBuild struct {
	ProjectName string `yaml:"project_name"`
}

type Hook struct {
	Event  string
	Refs   string
	Config string
}
