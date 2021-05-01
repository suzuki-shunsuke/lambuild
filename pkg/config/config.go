package config

import (
	"github.com/suzuki-shunsuke/matchfile-parser/matchfile"
)

type Config struct {
	Repositories []Repository
}

type Repository struct {
	Name      string
	Hooks     []Hook
	CodeBuild CodeBuild `yaml:"codebuild"`
}

type CodeBuild struct {
	ProjectName string `yaml:"project-name"`
}

type Hook struct {
	Event             []string
	Ref               string
	BaseRef           string `yaml:"base-ref"`
	Author            string
	Label             string
	RefConditions     []matchfile.Condition `yaml:"-"`
	BaseRefConditions []matchfile.Condition `yaml:"-"`
	AuthorConditions  []matchfile.Condition `yaml:"-"`
	LabelConditions   []matchfile.Condition `yaml:"-"`
	NoLabel           bool                  `yaml:"no-label"`
	Config            string
}
