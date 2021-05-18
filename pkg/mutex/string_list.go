package mutex

import (
	"sync"
)

type StringList struct {
	value []string
	mutex *sync.RWMutex
}

func NewStringList(list ...string) StringList {
	return StringList{
		value: list,
		mutex: &sync.RWMutex{},
	}
}

func (str *StringList) Get() []string {
	str.mutex.RLock()
	s := str.value
	str.mutex.RUnlock()
	return s
}

func (str *StringList) Set(value []string) {
	str.mutex.Lock()
	str.value = value
	str.mutex.Unlock()
}
