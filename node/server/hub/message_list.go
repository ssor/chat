package hub

import (
	"sync"

	"github.com/ssor/chat/node/server/communication"
)

type messageList struct {
	list map[string]*communication.Message
	lock *sync.Mutex
}

func newMessageList() *messageList {
	return &messageList{
		list: make(map[string]*communication.Message),
		lock: &sync.Mutex{},
	}
}

func (ml messageList) add(msg *communication.Message) {
	ml.lock.Lock()
	defer ml.lock.Unlock()
	ml.list[msg.GetID()] = msg
}

func (ml messageList) getMessageContent(id string) []byte {
	msg := ml.find(id)
	if msg != nil {
		return msg.GetContent()
	}
	return nil
}

func (ml messageList) find(id string) *communication.Message {
	ml.lock.Lock()
	defer ml.lock.Unlock()
	if msg, ok := ml.list[id]; ok {
		return msg
	}
	return nil
}
