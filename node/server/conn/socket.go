package conn

import "time"

type Socket interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	SetWriteDeadline(t time.Time) error
	Close() error
}
