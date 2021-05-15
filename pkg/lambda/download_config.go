package lambda

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/go-github/v35/github"
	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"gopkg.in/yaml.v2"
)

// getConfigFromRepo gets the configuration file from the target repository
func (handler *Handler) getConfigFromRepo(ctx context.Context, logE *logrus.Entry, data *domain.Data, hook config.Hook) ([]bspec.Buildspec, error) {
	// get the configuration file from the target repository
	if hook.Config == "" {
		// set the default value
		hook.Config = "lambuild.yaml"
	}
	file, files, _, err := handler.GitHub.Repositories.GetContents(ctx, data.Repository.Owner, data.Repository.Name, hook.Config, &github.RepositoryContentGetOptions{Ref: data.Ref})
	if err != nil {
		logE.WithFields(logrus.Fields{
			"path": hook.Config,
		}).WithError(err).Error("get configuration files by GitHub API")
		return nil, fmt.Errorf("get configuration files by GitHub API: %w", err)
	}
	if file != nil {
		files = []*github.RepositoryContent{file}
	}
	specs := make([]bspec.Buildspec, len(files))
	for i, file := range files {
		filePath := file.GetPath()
		ext := filepath.Ext(filePath)
		if ext != "yaml" && ext != "yml" {
			continue
		}
		content, err := file.GetContent()
		if err != nil {
			return nil, fmt.Errorf("get a content: %w", err)
		}
		if content == "" {
			f, _, _, err := handler.GitHub.Repositories.GetContents(ctx, data.Repository.Owner, data.Repository.Name, filePath, &github.RepositoryContentGetOptions{Ref: data.Ref})
			if err != nil {
				logE.WithFields(logrus.Fields{
					"path": filePath,
				}).WithError(err).Error("get a configuration file by GitHub API")
				return nil, fmt.Errorf("get a configuration file by GitHub API: %w", err)
			}
			cnt, err := f.GetContent()
			if err != nil {
				return nil, fmt.Errorf("get a content: %w", err)
			}
			if cnt == "" {
				return nil, fmt.Errorf("a content is empty (%s)", filePath)
			}
			content = cnt
		}

		m := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(content), &m); err != nil {
			return nil, fmt.Errorf("unmarshal a buildspec to map: %w", err)
		}

		buildspec := bspec.Buildspec{}
		if err := yaml.Unmarshal([]byte(content), &buildspec); err != nil {
			return nil, fmt.Errorf("unmarshal a buildspec: %w", err)
		}
		buildspec.Map = m
		specs[i] = buildspec
	}
	return specs, nil
}
