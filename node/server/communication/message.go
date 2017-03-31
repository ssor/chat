package communication

import (
	"encoding/json"
	"fmt"
	"time"
)

// type messageList []*Message

// Message format for communication with client
type Message struct {
	Protocal     int    `json:"protocol"`
	ID           string `json:"uid"`
	Name         string `json:"name"`
	Content      string `json:"content"`
	Timestamp    int64  `json:"timestamp"`
	MessageID    string `json:"messageID"`
	MessageBytes []byte `json:"-"`
}

// Fill add user's info to msg
func (message *Message) Fill(id, name string) error {
	ns := time.Now().UnixNano()
	message.Timestamp = ns
	message.ID = id
	message.Name = name
	message.MessageID = fmt.Sprintf("%d%s%d", message.Protocal, id, ns)

	bs, err := json.Marshal(message)
	if err != nil {
		return err
	}
	message.MessageBytes = bs
	return nil
}

// NewMessage init a message
func NewMessage(proto int, id, name, content string) (*Message, error) {
	ns := time.Now().UnixNano()
	m := &Message{
		Protocal:  proto,
		ID:        id,
		Name:      name,
		Content:   content,
		Timestamp: ns,
		MessageID: fmt.Sprintf("%d%s%d", proto, id, ns),
	}
	bs, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	m.MessageBytes = bs
	return m, nil
}

// GetID returns message's unique ID
func (message *Message) GetID() string {
	return message.MessageID
}

// GetContent returns data for binary
func (message *Message) GetContent() []byte {
	return message.MessageBytes
}

// GetProtoType returns message's protocol
func (message *Message) GetProtoType() int {
	return message.Protocal
}

//String use
func (message *Message) String() string {
	return fmt.Sprintf("protocol: %d id: %s  content: %s", message.Protocal, message.ID, message.Content)
}

// func newLoginMessage(u UserInfo) (*message, error) {
// 	return newMessage(protoLogin, u.GetUserID(), u.GetUserName(), "")
// }

// func newLogoutMessage(u UserInfo) (*message, error) {
// 	return newMessage(protoLogout, u.GetUserID(), u.GetUserName(), "")
// }

// NewTextMessage wrap for new text message
func NewTextMessage(id, name, content string) (*Message, error) {
	return NewMessage(ProtoText, id, name, content)
}

// NewImageMessage wrap for image message
func NewImageMessage(userID, userName, name string) (*Message, error) {
	return NewMessage(ProtoImage, userID, userName, name)
}

// NewAudioMessage wrap for audio message
func NewAudioMessage(userID, userName, name string) (*Message, error) {
	return NewMessage(ProtoAudio, userID, userName, name)
}

// func newCloseNotifyMessage() (*message, error) {
// 	return newMessage(protoClose, "", "", "")
// }
