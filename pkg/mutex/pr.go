package mutex

import (
	"sync"

	"github.com/google/go-github/v35/github"
)

type PR struct {
	value *github.PullRequest
	mutex *sync.RWMutex
}

func NewPR() PR {
	return PR{
		mutex: &sync.RWMutex{},
	}
}

func (mutex *PR) Get() *github.PullRequest {
	mutex.mutex.RLock()
	s := mutex.value
	mutex.mutex.RUnlock()
	return s
}

func (mutex *PR) Set(value *github.PullRequest) {
	mutex.mutex.Lock()
	mutex.value = value
	mutex.mutex.Unlock()
}
