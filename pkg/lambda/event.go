package lambda

import (
	"context"
	"fmt"
	"regexp"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/google/go-github/v35/github"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
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

type Data struct {
	Event             Event
	PullRequest       PullRequest
	Repository        Repository
	HeadCommitMessage string
	SHA               string
	Ref               string
	GitHub            *github.Client
	Commit            *github.Commit
	Lambuild          bspec.Lambuild
}

func runExpr(prog *vm.Program, data *Data) (interface{}, error) {
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

func setExprFuncs(env map[string]interface{}) map[string]interface{} {
	env["regexp"] = map[string]interface{}{
		"match": func(pattern, s string) bool {
			f, err := regexp.MatchString(pattern, s)
			if err != nil {
				panic(err)
			}
			return f
		},
	}
	return env
}

func (data *Data) CommitMessage() string {
	if data.HeadCommitMessage == "" {
		data.HeadCommitMessage = data.GetCommit().GetMessage()
	}
	return data.HeadCommitMessage
}

func (data *Data) GetCommit() *github.Commit {
	if data.Commit == nil {
		commit, _, err := data.GitHub.Git.GetCommit(context.Background(), data.Repository.Owner, data.Repository.Name, data.SHA)
		if err != nil {
			panic(err)
		}
		data.Commit = commit
	}
	return data.Commit
}

func (data *Data) GetPRNumber() int {
	if data.PullRequest.Number == 0 {
		if data.PullRequest.PullRequest != nil {
			data.PullRequest.Number = data.PullRequest.PullRequest.GetNumber()
			return data.PullRequest.Number
		}
		n, err := getPRNumber(context.Background(), data.Repository.Owner, data.Repository.Name, data.SHA, data.GitHub)
		if err != nil {
			panic(err)
		}
		data.PullRequest.Number = n
	}
	return data.PullRequest.Number
}

func (data *Data) GetPR() *github.PullRequest {
	if data.PullRequest.PullRequest == nil {
		pr, _, err := data.GitHub.PullRequests.Get(context.Background(), data.Repository.Owner, data.Repository.Name, data.GetPRNumber())
		if err != nil {
			panic(err)
		}
		data.PullRequest.PullRequest = pr
	}
	return data.PullRequest.PullRequest
}

func (data *Data) GetPRFiles() []*github.CommitFile {
	if data.PullRequest.Files == nil {
		files, _, err := getPRFiles(context.Background(), data.GitHub, data.Repository.Owner, data.Repository.Name, data.GetPRNumber(), data.GetPR().GetChangedFiles())
		if err != nil {
			panic(err)
		}
		data.PullRequest.Files = files
	}
	return data.PullRequest.Files
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
