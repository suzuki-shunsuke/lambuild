package domain

import (
	"context"
	"fmt"

	"github.com/google/go-github/v35/github"
)

func extractLabelNames(labels []*github.Label) []string {
	labelNames := make([]string, len(labels))
	for i, label := range labels {
		labelNames[i] = label.GetName()
	}
	return labelNames
}

func extractPRFileNames(files []*github.CommitFile) []string {
	prFileNames := make(map[string]struct{}, len(files))
	for _, file := range files {
		prFileNames[file.GetFilename()] = struct{}{}
		prevFileName := file.GetPreviousFilename()
		if prevFileName != "" {
			prFileNames[prevFileName] = struct{}{}
		}
	}
	arr := make([]string, 0, len(prFileNames))
	for k := range prFileNames {
		arr = append(arr, k)
	}
	return arr
}

func getPRNumber(ctx context.Context, owner, repo, sha string, client *github.Client) (int, error) {
	prs, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return 0, fmt.Errorf("list pull requests with a commit: %w", err)
	}
	if len(prs) == 0 {
		return 0, nil
	}
	return prs[0].GetNumber(), nil
}

const maxPerPage = 100

func getPRFiles(ctx context.Context, client *github.Client, owner, repo string, prNum, fileSize int) ([]*github.CommitFile, *github.Response, error) {
	ret := []*github.CommitFile{}
	if fileSize == 0 {
		return nil, nil, nil
	}
	n := (fileSize / maxPerPage) + 1
	var gResp *github.Response
	for i := 1; i <= n; i++ {
		opts := &github.ListOptions{
			Page:    i,
			PerPage: maxPerPage,
		}
		files, resp, err := client.PullRequests.ListFiles(ctx, owner, repo, prNum, opts)
		if err != nil {
			return files, resp, fmt.Errorf("get pull request files (page: %d, per_page: %d): %w", opts.Page, opts.PerPage, err)
		}
		gResp = resp
		ret = append(ret, files...)
		if len(files) != maxPerPage {
			return ret, gResp, nil
		}
	}

	return ret, gResp, nil
}
