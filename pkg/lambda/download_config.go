package lambda

import (
	"context"
	"fmt"

	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"gopkg.in/yaml.v2"
)

// getConfigFromRepo gets the configuration file from the target repository
func (handler *Handler) getConfigFromRepo(ctx context.Context, logE *logrus.Entry, data *domain.Data, hook config.Hook) (bspec.Buildspec, error) {
	buildspec := bspec.Buildspec{}
	// get the configuration file from the target repository
	if hook.Config == "" {
		// set the default value
		hook.Config = "lambuild.yaml"
	}
	file, _, _, err := handler.GitHub.Repositories.GetContents(ctx, data.Repository.Owner, data.Repository.Name, hook.Config, &github.RepositoryContentGetOptions{Ref: data.Ref})
	if err != nil {
		logE.WithFields(logrus.Fields{
			"path": hook.Config,
		}).WithError(err).Error("")
		return buildspec, fmt.Errorf("get a configuration file by GitHub API: %w", err)
	}
	content, err := file.GetContent()
	if err != nil {
		return buildspec, fmt.Errorf("get a content: %w", err)
	}

	m := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(content), &m); err != nil {
		return buildspec, fmt.Errorf("unmarshal a buildspec to map: %w", err)
	}

	if err := yaml.Unmarshal([]byte(content), &buildspec); err != nil {
		return buildspec, fmt.Errorf("unmarshal a buildspec: %w", err)
	}
	data.Lambuild = buildspec.Lambuild
	buildspec.Lambuild = bspec.Lambuild{}
	buildspec.Map = m
	return buildspec, nil
}