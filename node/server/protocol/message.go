package protocol

import (
	"encoding/json"
	"fmt"
	"time"
)

var (
	// protoUnknown = -1
	// protoLogin  = 0
	// protoLogout = 1

	// ProtoText text
	ProtoText = 2
	// ProtoImage image upload
	ProtoImage = 3
	// ProtoAudio audio upload
	ProtoAudio = 4
	// ProtoShare share some thing
	ProtoShare = 5
	// ProtoReply reply from client
	ProtoReply = 6
	// ProtoBot bot chat
	ProtoBot = 888
	// ProtoCloseLoginOnOtherDevice server close conn and do not want client to reconnect in short time
	ProtoCloseLoginOnOtherDevice = "loginOnOtherDevice"
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

func (message *Message) fill(id, name string) error {
	ns := time.Now().UnixNano()
	message.Timestamp = ns
	message.ID = id
	message.Name = name
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

// func newTextMessage(u *User, content string) *message {
// 	return newMessage(protoText, u.ID, u.Name, content)
// }
//

// func NewImageMessage(userID, userName, name string) (*Message, error) {
// 	// return newMessage(protoImage, u.ID, u.Name, ResourceDirMapList[imageDir].FullURL()+name)
// 	return NewMessage(ProtoImage, userID, userName, name)
// }

// func NewAudioMessage(userID, userName, name string) (*Message, error) {
// 	// return newMessage(protoAudio, u.ID, u.Name, ResourceDirMapList[audioDir].FullURL()+name)
// 	return NewMessage(ProtoAudio, userID, userName, name)
// }

// func newCloseNotifyMessage() (*message, error) {
// 	return newMessage(protoClose, "", "", "")
// }
