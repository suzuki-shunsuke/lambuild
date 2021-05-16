package lambda

import (
	"context"

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
	var cmt string
	s, renderErr := handler.Config.ErrorNotificationTemplate.Execute(map[string]interface{}{
		"Error": e,
	})
	if renderErr != nil {
		logE.WithError(renderErr).Error("render a comment to send it to the pull request")
		cmt = "lambuild failed to procceed the request: " + e.Error()
	} else {
		cmt = s
	}

	if prNumber == 0 {
		// send a comment to commit
		if cmtErr := handler.GitHub.CreateCommitComment(ctx, repoOwner, repoName, sha, cmt); cmtErr != nil {
			logE.WithError(cmtErr).Error("send a comment to the commit")
		}
		logE.Info("send a comment to the commit")
		return
	}

	// send a comment to pull request
	if cmtErr := handler.GitHub.CreatePRComment(ctx, repoOwner, repoName, prNumber, cmt); cmtErr != nil {
		logE.WithError(cmtErr).Error("send a comment to the pull request")
	}
	logE.Info("send a comment to the pull request")
}
