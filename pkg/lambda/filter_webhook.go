package lambda

import (
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
)

// getRepo returns the configuration of given repository name.
// If no configuration is found, the second returned value is false.
func getRepo(repos []config.Repository, repoName string) (config.Repository, bool) {
	for _, repo := range repos {
		if repo.Name == repoName {
			return repo, true
		}
	}
	return config.Repository{}, false
}

// matchHook returns true if data matches hook's condition.
func matchHook(data *domain.Data, hook config.Hook) (bool, error) {
	if hook.If.Empty() {
		return true, nil
	}
	return hook.If.Run(data.Convert()) //nolint:wrapcheck
}

// getHook returns a hook configuration which data matches.
// If data doesn't match any configuration, the second returned value is false.
func getHook(data *domain.Data, repo config.Repository) (config.Hook, bool, error) {
	for _, hook := range repo.Hooks {
		f, err := matchHook(data, hook)
		if err != nil {
			return config.Hook{}, false, err
		}
		if f {
			return hook, true, nil
		}
	}
	return config.Hook{}, false, nil
}
