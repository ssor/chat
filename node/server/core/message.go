package core

import (
	"sync"
)

type message interface {
	GetID() string
	GetContent() []byte
}

type messageList struct {
	list map[string]message
	lock *sync.Mutex
}

func newMessageList() *messageList {
	return &messageList{
		list: make(map[string]message),
		lock: &sync.Mutex{},
	}
}

func (ml messageList) add(msg message) {
	ml.lock.Lock()
	defer ml.lock.Unlock()
	ml.list[msg.GetID()] = msg
}
