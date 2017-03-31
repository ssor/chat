package user

import "xsbPro/chat/node/server/communication"

type messageStore interface {
	GetMessage(string) *communication.Message
	PopNewMessage(*communication.Message)
}
