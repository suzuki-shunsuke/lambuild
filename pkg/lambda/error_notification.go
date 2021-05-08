package lambda

import (
	"bytes"
	"context"

	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
)

// sendErrorNotificaiton sends a comment to GitHub PullRequest or commit to notify an error.
// If prNumber isn't zero a comment is sent to the pull reqquest.
// If prNumber is zero, which means the event isn't associated with any pull request, a comment is sent to a comment.
func (handler *Handler) sendErrorNotificaiton(ctx context.Context, e error, repoOwner, repoName string, prNumber int, sha string) {
	logE := logrus.WithFields(logrus.Fields{
		"original_error": e,
		"repo_owner":     repoOwner,
		"repo_name":      repoName,
		"pr_number":      prNumber,
		"sha":            sha,
	})

	// generate a comment
	buf := &bytes.Buffer{}
	var cmt string
	if renderErr := handler.Config.ErrorNotificationTemplate.Execute(buf, map[string]interface{}{
		"Error": e,
	}); renderErr != nil {
		logE.WithError(renderErr).Error("render a comment to send it to the pull request")
		cmt = "lambuild failed to procceed the request: " + e.Error()
	} else {
		cmt = buf.String()
	}

	if prNumber == 0 {
		// send a comment to commit
		if _, _, cmtErr := handler.GitHub.Repositories.CreateComment(ctx, repoOwner, repoName, sha, &github.RepositoryComment{
			Body: github.String(cmt),
		}); cmtErr != nil {
			logE.WithError(cmtErr).Error("send a comment to the commit")
		}
		logE.Info("send a comment to the commit")
		return
	}

	// send a comment to pull request
	if _, _, cmtErr := handler.GitHub.Issues.CreateComment(ctx, repoOwner, repoName, prNumber, &github.IssueComment{
		Body: github.String(cmt),
	}); cmtErr != nil {
		logE.WithError(cmtErr).Error("send a comment to the pull request")
	}
	logE.Info("send a comment to the pull request")
}
