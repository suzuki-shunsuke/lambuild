package mutex

import (
	"sync"
)

type String struct {
	value string
	mutex *sync.RWMutex
}

func NewString(val string) String {
	return String{
		value: val,
		mutex: &sync.RWMutex{},
	}
}

func (str *String) Get() string {
	str.mutex.RLock()
	s := str.value
	str.mutex.RUnlock()
	return s
}

func (str *String) Set(value string) {
	str.mutex.Lock()
	str.value = value
	str.mutex.Unlock()
}
