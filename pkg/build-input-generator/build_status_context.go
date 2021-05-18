package generator

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func setBuildStatusContext(contxt template.Template, data *domain.Data, input *codebuild.StartBuildInput) error {
	s, err := getBuildStatusContext(contxt, data)
	if err != nil || s == "" {
		return err
	}
	input.BuildStatusConfigOverride = &codebuild.BuildStatusConfig{
		Context: aws.String(s),
	}
	return nil
}

func getBuildStatusContext(tpl template.Template, data *domain.Data) (string, error) {
	s, err := tpl.Execute(map[string]interface{}{
		"event": data.Event,
		"pr":    data.PullRequest,
		"repo":  data.Repository,
		"sha":   data.SHA,
		"ref":   data.Ref,
	})
	if err != nil {
		return "", fmt.Errorf("render a build status context: %w", err)
	}
	return s, nil
}
