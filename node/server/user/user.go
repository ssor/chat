package user

import (
	"context"
	"encoding/json"
	"sync"
	"time"
	"xsbPro/chat/node/server/communication"
	conn "xsbPro/chat/node/server/connection"
	"xsbPro/log"
)

var maxMessageCountCache = 1024

// NewUser init a user in hub
func NewUser(d detail, store messageStore) *User {
	return &User{
		detail:         d,
		messageStore:   store,
		MessageRecords: []string{},
		lastActiveTime: time.Now(),
		lockConn:       &sync.Mutex{},
		lockMessage:    &sync.Mutex{},
	}
}

// User maps the user of client
// it holds the messages the user should receive, and use conn to send it to client
// it also make sure messages is received by client
type User struct {
	detail         detail
	conn           *conn.Connection
	MessageRecords []string
	messageStore   messageStore
	lastActiveTime time.Time
	lockConn       *sync.Mutex
	lockMessage    *sync.Mutex // lock for dispose of message
}

// func (u *User) SetMessagePopHandler(func(*protocol.Message) error) {

// }

// GetID return id for this user
func (u *User) GetID() string {
	return u.detail.GetID()
}

// GetName return name for this user
func (u *User) GetName() string {
	return u.detail.GetName()
}

// SetConn wraps for connection SetConn
func (u *User) SetConn(skt conn.Socket, cancelSocket context.CancelFunc) {
	u.lockConn.Lock()
	defer u.lockConn.Unlock()
	if u.conn == nil {
		return
	}
	u.conn.SetConn(skt)
}

// SendMessage send the first message in cache to client
// message is sent one by one,
// one message sent, server wait for it's reply; if received reply, send success, and the next one
// if no reply, the same message will be sent again
func (u *User) SendMessage() {
	conn := u.conn
	if conn == nil {
		return
	}

	u.lockMessage.Lock()
	defer u.lockMessage.Unlock()

	if len(u.MessageRecords) <= 0 {
		return
	}

	msgID := u.MessageRecords[0]
	msg := u.messageStore.GetMessage(msgID)
	if msg != nil {
		conn.Send(msg.GetContent())
		// if err != nil {
		// 	u.releaseConn()
		// }
	} else {
		// 该消息已经本体已经不存在, 发送记录中不应再保存
		go u.RemoveRecordCache(msgID)
	}
}

// RemoveRecordCache removes specified msg
func (u *User) RemoveRecordCache(msgID string) {
	u.lockMessage.Lock()
	defer u.lockMessage.Unlock()

	for index, id := range u.MessageRecords {
		if msgID == id {
			u.MessageRecords = removeStringFromSlice(u.MessageRecords, index)
			return
		}
	}
}

func removeStringFromSlice(slice []string, index int) []string {
	if index >= len(slice) {
		return slice[:]
	}
	return append(slice[:index], slice[index+1:]...)
}

// Release release resoures of user
func (u *User) Release() {
	u.lockConn.Lock()
	defer u.lockConn.Unlock()
	if u.conn != nil {
		u.conn.Close("", time.Second*1)
	}
	// u.conn = nil
}

// NewMessage new data upload from user's conn
func (u *User) NewMessage(mIn *communication.Message) (err error) {
	switch mIn.Protocal {
	case communication.ProtoShare, communication.ProtoText:
		err = mIn.Fill(u.detail.GetID(), u.detail.GetName())
		if err != nil {
			log.InfoF("read message error: %s", err)
			return
		}
		go u.messageStore.PopNewMessage(mIn)
		// err = u.hub.NewMessage(&mIn, nil)
		// if err != nil {
		// 	log.SysF("NewMessage error: %s", err.Error())
		// }
	case communication.ProtoReply:
		u.RemoveRecordCache(mIn.GetID())
		// err = u.hub.NewMessage(&mIn, []string{u.User.GetUserID()})
		// if err != nil {
		// 	log.SysF("NewMessage error: %s", err.Error())
		// }
	default:
	}
	return
}

// NewDataIn new data upload from user's conn
func (u *User) NewDataIn(data []byte) error {
	var mIn communication.Message
	err := json.Unmarshal(data, &mIn)
	// if err == nil && (mIn.Protocal == protoShare || mIn.Protocal == protoText) {
	if err == nil {
		return u.NewMessage(&mIn)
	}
	log.SysF("NewMessage error: %s %s", err.Error(), string(data))
	return err
}

// AddMessageToCache stores msg id to cache, then msges will be sent to client
func (u *User) AddMessageToCache(msgID string) {
	//如果为虚拟用户,那么不在线时不接收消息
	// if u.User.IsFake() && u.conn == nil {
	// 	return
	// }

	// r := newMessageRecord(m)
	msg := u.messageStore.GetMessage(msgID)
	switch msg.GetProtoType() {
	case communication.ProtoReply:
		u.MessageRecords = append([]string{msgID}, u.MessageRecords...)
	default:
		u.MessageRecords = append(u.MessageRecords, msgID)
	}

	if len(u.MessageRecords) > maxMessageCountCache {
		u.MessageRecords = u.MessageRecords[:maxMessageCountCache]
	}
}
