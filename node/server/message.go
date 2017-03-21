package server

import (
	"encoding/json"
	"fmt"
	// "chatServer/libs/log"
	"time"
)

var (
	// protoUnknown = -1
	// protoLogin  = 0
	// protoLogout = 1
	protoText  = 2
	protoImage = 3
	protoAudio = 4
	protoShare = 5
	protoReply = 6
	protoBot   = 888

	protoCloseLoginOnOtherDevice string = "loginOnOtherDevice" //server close conn and do not want client to reconnect in short time
)

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

func NewMessage(proto int, id, name, content string) (*Message, error) {
	ns := time.Now().UnixNano()
	m := &Message{
		Protocal:  proto,
		ID:        id,
		Name:      name,
		Content:   content,
		Timestamp: ns,
	}
	bs, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	m.MessageBytes = bs
	return m, nil
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
func NewImageMessage(u UserInfo, name string) (*Message, error) {
	// return newMessage(protoImage, u.ID, u.Name, ResourceDirMapList[imageDir].FullURL()+name)
	return NewMessage(protoImage, u.GetUserID(), u.GetUserName(), name)
}

func NewAudioMessage(u UserInfo, name string) (*Message, error) {
	// return newMessage(protoAudio, u.ID, u.Name, ResourceDirMapList[audioDir].FullURL()+name)
	return NewMessage(protoAudio, u.GetUserID(), u.GetUserName(), name)
}

// func newCloseNotifyMessage() (*message, error) {
// 	return newMessage(protoClose, "", "", "")
// }
