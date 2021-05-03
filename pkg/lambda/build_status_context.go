package lambda

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
)

func (handler *Handler) setBuildStatusContext(data *Data, input *codebuild.StartBuildInput) error {
	if handler.BuildStatusContext == nil {
		return nil
	}
	buf := &bytes.Buffer{}
	if err := handler.BuildStatusContext.Execute(buf, map[string]interface{}{
		"event":         data.Event,
		"pr":            data.PullRequest,
		"repo":          data.Repository,
		"sha":           data.SHA,
		"ref":           data.Ref,
		"commit":        data.GetCommit,
		"commitMessage": data.CommitMessage,
	}); err != nil {
		return fmt.Errorf("render a build status context: %w", err)
	}
	input.BuildStatusConfigOverride = &codebuild.BuildStatusConfig{
		Context: aws.String(buf.String()),
	}
	return nil
}
