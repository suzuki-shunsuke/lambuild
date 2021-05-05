package lambda

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
)

func (handler *Handler) setBuildStatusContext(data *Data, input *codebuild.StartBuildInput) error {
	s, err := getBuildStatusContext(handler.Config.BuildStatusContext, data)
	if err != nil || s == "" {
		return err
	}
	input.BuildStatusConfigOverride = &codebuild.BuildStatusConfig{
		Context: aws.String(s),
	}
	return nil
}

func getBuildStatusContext(tpl *template.Template, data *Data) (string, error) {
	if tpl == nil {
		return "", nil
	}
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, map[string]interface{}{
		"event": data.Event,
		"pr":    data.PullRequest,
		"repo":  data.Repository,
		"sha":   data.SHA,
		"ref":   data.Ref,
	}); err != nil {
		return "", fmt.Errorf("render a build status context: %w", err)
	}
	return buf.String(), nil
}
