package user

import "github.com/ssor/chat/node/server/communication"

type messageStore interface {
	GetMessage(string) *communication.Message
	PopNewMessage(*communication.Message)
}
