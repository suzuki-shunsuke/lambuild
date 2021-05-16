package mutex

import (
	"sync"

	"github.com/google/go-github/v35/github"
)

type CommitFiles struct {
	value []*github.CommitFile
	mutex *sync.RWMutex
}

func NewCommitFiles() CommitFiles {
	return CommitFiles{
		mutex: &sync.RWMutex{},
	}
}

func (mutex *CommitFiles) Get() []*github.CommitFile {
	mutex.mutex.RLock()
	s := mutex.value
	mutex.mutex.RUnlock()
	return s
}

func (mutex *CommitFiles) Set(value []*github.CommitFile) {
	mutex.mutex.Lock()
	mutex.value = value
	mutex.mutex.Unlock()
}
