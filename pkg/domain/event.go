package domain

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/google/go-github/v35/github"
)

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
	Files            []*github.CommitFile
	Number           int
}

type Repository struct {
	FullName string
	Owner    string
	Name     string
}

// Data contains data which is referred in expression engine and template engine.
// To reduce unneeded HTTP API call, in Data's many functions API isn't called until the API call is really needed, and the result is cached in the request scope.
type Data struct {
	Event             Event
	PullRequest       PullRequest
	Repository        Repository
	HeadCommitMessage string
	SHA               string
	Ref               string
	GitHub            *github.Client
	Commit            *github.Commit
}

func RunExpr(prog *vm.Program, data *Data) (interface{}, error) {
	result, err := expr.Run(prog, setExprFuncs(map[string]interface{}{
		"event":            data.Event,
		"repo":             data.Repository,
		"sha":              data.SHA,
		"ref":              data.Ref,
		"getCommit":        data.GetCommit,
		"getCommitMessage": data.CommitMessage,
		"getPR":            data.GetPR,
		"getPRNumber":      data.GetPRNumber,
		"getPRFiles":       data.GetPRFiles,
		"getPRFileNames":   data.GetPRFileNames,
		"getPRLabelNames":  data.GetPRLabelNames,
	}))
	if err != nil {
		return nil, fmt.Errorf("evaluate a expr's compiled program: %w", err)
	}
	return result, nil
}

func (data *Data) CommitMessage() string {
	if data.HeadCommitMessage == "" {
		data.HeadCommitMessage = data.GetCommit().GetMessage()
	}
	return data.HeadCommitMessage
}

func (data *Data) GetPRFileNames() []string {
	if data.PullRequest.ChangedFileNames == nil {
		data.PullRequest.ChangedFileNames = extractPRFileNames(data.GetPRFiles())
	}
	return data.PullRequest.ChangedFileNames
}

func (data *Data) GetPRLabelNames() []string {
	if data.PullRequest.LabelNames == nil {
		data.PullRequest.LabelNames = extractLabelNames(data.GetPR().Labels)
	}
	return data.PullRequest.LabelNames
}
