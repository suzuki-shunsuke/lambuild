package mutex

import (
	"sync"
)

type Int struct {
	value int
	mutex *sync.RWMutex
}

func NewInt() Int {
	return Int{
		mutex: &sync.RWMutex{},
	}
}

func (mutex *Int) Get() int {
	mutex.mutex.RLock()
	s := mutex.value
	mutex.mutex.RUnlock()
	return s
}

func (mutex *Int) Set(value int) {
	mutex.mutex.Lock()
	mutex.value = value
	mutex.mutex.Unlock()
}
