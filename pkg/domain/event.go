package domain

import (
	"context"

	"github.com/google/go-github/v35/github"
	"github.com/suzuki-shunsuke/lambuild/pkg/mutex"
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
	ChangedFileNames mutex.StringList
	LabelNames       mutex.StringList
	PullRequest      mutex.PR
	Files            mutex.CommitFiles
	Number           mutex.Int
}

func NewPullRequest() PullRequest {
	return PullRequest{
		ChangedFileNames: mutex.NewStringList(),
		LabelNames:       mutex.NewStringList(),
		PullRequest:      mutex.NewPR(),
		Files:            mutex.NewCommitFiles(),
		Number:           mutex.NewInt(),
	}
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
	HeadCommitMessage mutex.String
	SHA               string
	Ref               string
	GitHub            GitHub
	Commit            mutex.Commit
	AWS               AWSData
}

type AWSData struct {
	Region    string
	AccountID string
}

type GitHub interface {
	GetCommit(ctx context.Context, owner, repo, sha string) (*github.Commit, error)
	GetPR(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error)
	GetPRFiles(ctx context.Context, owner, repo string, number int, opt *github.ListOptions) ([]*github.CommitFile, error)
	GetPRsWithCommit(ctx context.Context, owner, repo string, sha string) ([]*github.PullRequest, error)
	GetContents(ctx context.Context, owner, repo, path, ref string) (*github.RepositoryContent, []*github.RepositoryContent, error)
	CreateCommitComment(ctx context.Context, owner, repo, sha, body string) error
	CreatePRComment(ctx context.Context, owner, repo string, number int, body string) error
}

func NewData() Data {
	return Data{
		Commit:            mutex.NewCommit(),
		HeadCommitMessage: mutex.NewString(""),
		PullRequest:       NewPullRequest(),
	}
}

func (data *Data) Convert() map[string]interface{} {
	return setExprFuncs(map[string]interface{}{
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
		"aws": map[string]interface{}{
			"region":     data.AWS.Region,
			"account_id": data.AWS.AccountID,
		},
	})
}

func (data *Data) CommitMessage() string {
	if msg := data.HeadCommitMessage.Get(); msg != "" {
		return msg
	}
	msg := data.GetCommit().GetMessage()
	data.HeadCommitMessage.Set(msg)
	return msg
}

func (data *Data) GetPRFileNames() []string {
	if val := data.PullRequest.ChangedFileNames.Get(); val != nil {
		return val
	}
	val := extractPRFileNames(data.GetPRFiles())
	data.PullRequest.ChangedFileNames.Set(val)
	return val
}

func (data *Data) GetPRLabelNames() []string {
	if val := data.PullRequest.LabelNames.Get(); val != nil {
		return val
	}
	val := extractLabelNames(data.GetPR().Labels)
	data.PullRequest.LabelNames.Set(val)
	return val
}
