package domain

import (
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
	ChangedFileNames StringListMutex
	LabelNames       StringListMutex
	PullRequest      PRMutex
	Files            CommitFilesMutex
	Number           IntMutex
}

func NewPullRequest() PullRequest {
	return PullRequest{
		ChangedFileNames: NewStringListMutex(),
		LabelNames:       NewStringListMutex(),
		PullRequest:      NewPRMutex(),
		Files:            NewCommitFilesMutex(),
		Number:           NewIntMutex(),
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
	HeadCommitMessage StringMutex
	SHA               string
	Ref               string
	GitHub            *github.Client
	Commit            CommitMutex
}

func NewData() Data {
	return Data{
		Commit:            NewCommitMutex(),
		HeadCommitMessage: NewStringMutex(""),
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
