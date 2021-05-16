package domain

import (
	"sync"

	"github.com/google/go-github/v35/github"
)

type StringMutex struct {
	value string
	mutex *sync.RWMutex
}

func NewStringMutex(val string) StringMutex {
	return StringMutex{
		value: val,
		mutex: &sync.RWMutex{},
	}
}

func (str *StringMutex) Get() string {
	str.mutex.RLock()
	s := str.value
	str.mutex.RUnlock()
	return s
}

func (str *StringMutex) Set(value string) {
	str.mutex.Lock()
	str.value = value
	str.mutex.Unlock()
}

type StringListMutex struct {
	value []string
	mutex *sync.RWMutex
}

func NewStringListMutex(list ...string) StringListMutex {
	return StringListMutex{
		value: list,
		mutex: &sync.RWMutex{},
	}
}

func (str *StringListMutex) Get() []string {
	str.mutex.RLock()
	s := str.value
	str.mutex.RUnlock()
	return s
}

func (str *StringListMutex) Set(value []string) {
	str.mutex.Lock()
	str.value = value
	str.mutex.Unlock()
}

type PRMutex struct {
	value *github.PullRequest
	mutex *sync.RWMutex
}

func NewPRMutex() PRMutex {
	return PRMutex{
		mutex: &sync.RWMutex{},
	}
}

func (mutex *PRMutex) Get() *github.PullRequest {
	mutex.mutex.RLock()
	s := mutex.value
	mutex.mutex.RUnlock()
	return s
}

func (mutex *PRMutex) Set(value *github.PullRequest) {
	mutex.mutex.Lock()
	mutex.value = value
	mutex.mutex.Unlock()
}

type IntMutex struct {
	value int
	mutex *sync.RWMutex
}

func NewIntMutex() IntMutex {
	return IntMutex{
		mutex: &sync.RWMutex{},
	}
}

func (mutex *IntMutex) Get() int {
	mutex.mutex.RLock()
	s := mutex.value
	mutex.mutex.RUnlock()
	return s
}

func (mutex *IntMutex) Set(value int) {
	mutex.mutex.Lock()
	mutex.value = value
	mutex.mutex.Unlock()
}

type CommitFilesMutex struct {
	value []*github.CommitFile
	mutex *sync.RWMutex
}

func NewCommitFilesMutex() CommitFilesMutex {
	return CommitFilesMutex{
		mutex: &sync.RWMutex{},
	}
}

func (mutex *CommitFilesMutex) Get() []*github.CommitFile {
	mutex.mutex.RLock()
	s := mutex.value
	mutex.mutex.RUnlock()
	return s
}

func (mutex *CommitFilesMutex) Set(value []*github.CommitFile) {
	mutex.mutex.Lock()
	mutex.value = value
	mutex.mutex.Unlock()
}

type CommitMutex struct {
	value *github.Commit
	mutex *sync.RWMutex
}

func NewCommitMutex() CommitMutex {
	return CommitMutex{
		mutex: &sync.RWMutex{},
	}
}

func (mutex *CommitMutex) Get() *github.Commit {
	mutex.mutex.RLock()
	s := mutex.value
	mutex.mutex.RUnlock()
	return s
}

func (mutex *CommitMutex) Set(value *github.Commit) {
	mutex.mutex.Lock()
	mutex.value = value
	mutex.mutex.Unlock()
}
