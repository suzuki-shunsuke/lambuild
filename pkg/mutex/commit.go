package mutex

import (
	"sync"

	"github.com/google/go-github/v36/github"
)

type Commit struct {
	value *github.Commit
	mutex *sync.RWMutex
}

func NewCommit() Commit {
	return Commit{
		mutex: &sync.RWMutex{},
	}
}

func (mutex *Commit) Get() *github.Commit {
	mutex.mutex.RLock()
	s := mutex.value
	mutex.mutex.RUnlock()
	return s
}

func (mutex *Commit) Set(value *github.Commit) {
	mutex.mutex.Lock()
	mutex.value = value
	mutex.mutex.Unlock()
}
