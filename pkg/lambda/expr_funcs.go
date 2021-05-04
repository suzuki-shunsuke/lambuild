package lambda

import (
	"context"
	"regexp"

	"github.com/google/go-github/v35/github"
)

// Functions and methods in this file is called at antonmedv/expr's program.
// Basically we shouldn't use `panic` for error handling,
// but in antonmedv/expr's program, `panic` is recommended.
// Please see the author's message.
//
// https://github.com/antonmedv/expr/issues/66#issuecomment-526461353
// > Hi, this is a good question. I think I forgot to mention this in docs. Expr expects functions to return a single value. To notify an error, you can use panic. Expr uses recover.
// > For example, if your function going to divide by zero 1/0 Expr will also work correctly as it will panic.

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
