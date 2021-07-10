package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
}

func New(ctx context.Context, token string) Client {
	return Client{
		client: github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))),
	}
}

func (client *Client) GetCommit(ctx context.Context, owner, repo, sha string) (*github.Commit, error) {
	commit, _, err := client.client.Git.GetCommit(ctx, owner, repo, sha)
	if err != nil {
		return nil, fmt.Errorf("get a commit by GitHub API: %w", err)
	}
	return commit, nil
}

func (client *Client) GetPR(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := client.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("get a pull request by GitHub API: %w", err)
	}
	return pr, nil
}

func (client *Client) GetPRFiles(ctx context.Context, owner, repo string, number int, opt *github.ListOptions) ([]*github.CommitFile, error) {
	files, _, err := client.client.PullRequests.ListFiles(ctx, owner, repo, number, opt)
	if err != nil {
		return nil, fmt.Errorf("get pull request files by GitHub API: %w", err)
	}
	return files, nil
}

func (client *Client) GetPRsWithCommit(ctx context.Context, owner, repo string, sha string) ([]*github.PullRequest, error) {
	prs, _, err := client.client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("get pull requests with commit by GitHub API: %w", err)
	}
	return prs, nil
}

func (client *Client) GetContents(ctx context.Context, owner, repo, path, ref string) (*github.RepositoryContent, []*github.RepositoryContent, error) {
	file, files, _, err := client.client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: ref})
	if err != nil {
		return nil, nil, fmt.Errorf("get pull requests with commit by GitHub API: %w", err)
	}
	return file, files, nil
}

func (client *Client) CreateCommitComment(ctx context.Context, owner, repo, sha, body string) error {
	if _, _, err := client.client.Repositories.CreateComment(ctx, owner, repo, sha, &github.RepositoryComment{
		Body: github.String(body),
	}); err != nil {
		return fmt.Errorf("create a commit comment by GitHub API: %w", err)
	}
	return nil
}

func (client *Client) CreatePRComment(ctx context.Context, owner, repo string, number int, body string) error {
	if _, _, err := client.client.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{
		Body: github.String(body),
	}); err != nil {
		return fmt.Errorf("create a pull request comment by GitHub API: %w", err)
	}
	return nil
}
