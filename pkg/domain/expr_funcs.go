package domain

import (
	"context"
	"reflect"

	"github.com/google/go-github/v36/github"
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
	m := make(map[string]interface{}, len(env)+1)
	for k, v := range env {
		m[k] = v
	}
	m["util"] = map[string]interface{}{
		"value": Value,
	}
	return m
}

func Value(ptr interface{}) interface{} {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return ptr
	}
	return v.Elem().Interface()
}

func (data *Data) GetCommit() *github.Commit {
	if cmt := data.Commit.Get(); cmt != nil {
		return cmt
	}
	commit, err := data.GitHub.GetCommit(context.Background(), data.Repository.Owner, data.Repository.Name, data.SHA)
	if err != nil {
		panic(err)
	}
	data.Commit.Set(commit)
	return commit
}

func (data *Data) GetPRNumber() int {
	if number := data.PullRequest.Number.Get(); number != 0 {
		return number
	}

	if pr := data.PullRequest.PullRequest.Get(); pr != nil {
		number := pr.GetNumber()
		data.PullRequest.Number.Set(number)
		return number
	}

	n, err := getPRNumber(context.Background(), data.Repository.Owner, data.Repository.Name, data.SHA, data.GitHub)
	if err != nil {
		panic(err)
	}
	data.PullRequest.Number.Set(n)
	return n
}

func (data *Data) GetPR() *github.PullRequest {
	pr := data.PullRequest.PullRequest.Get()
	if pr == nil {
		p, err := data.GitHub.GetPR(context.Background(), data.Repository.Owner, data.Repository.Name, data.GetPRNumber())
		if err != nil {
			panic(err)
		}
		pr = p
		data.PullRequest.PullRequest.Set(pr)
	}
	return pr
}

func (data *Data) GetPRFiles() []*github.CommitFile {
	if files := data.PullRequest.Files.Get(); files != nil {
		return files
	}
	files, err := getPRFiles(context.Background(), data.GitHub, data.Repository.Owner, data.Repository.Name, data.GetPRNumber(), data.GetPR().GetChangedFiles())
	if err != nil {
		panic(err)
	}
	data.PullRequest.Files.Set(files)
	return files
}
