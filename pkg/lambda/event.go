package lambda

import (
	"github.com/google/go-github/v35/github"
)

type Data struct {
	Event             Event
	PullRequest       PullRequest
	Repository        Repository
	HeadCommitMessage string
	SHA               string
	Ref               string
}

type Event struct {
	Body    string  `json:"body"`
	Headers Headers `json:"headers"`

	Payload interface{} `json:"-"`
}

type Headers struct {
	Event     string `json:"x-github-event"`
	Delivery  string `json:"x-github-delivery"`
	Signature string `json:"x-hub-signature-256"`
}

type PullRequest struct {
	ChangedFileNames []string
	LabelNames       []string
	PullRequest      *github.PullRequest
}

type Repository struct {
	FullName string
	Owner    string
	Name     string
}
