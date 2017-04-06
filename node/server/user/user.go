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

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var maxMessageCountCache = 1024

// NewUser init a user in hub
func NewUser(d detail, store messageStore) *User {
	u := &User{
		detail:         d,
		messageStore:   store,
		MessageRecords: []string{},
		lastActiveTime: time.Now(),
		lockConn:       &sync.Mutex{},
		lockMessage:    &sync.Mutex{},
		replys:         replyContentList{},
	}
	u.conn = conn.NewConnection(u.detail.GetID(), u, pingPeriod, writeWait)
	return u
}

// User maps the user of client
// it holds the messages the user should receive, and use conn to send it to client
// it also make sure messages is received by client
type User struct {
	detail         detail
	conn           *conn.Connection
	MessageRecords []string         // 待发送的消息的 ID
	replys         replyContentList // 待发送的已接收消息的回执
	messageStore   messageStore
	lastActiveTime time.Time
	lockConn       *sync.Mutex
	lockMessage    *sync.Mutex // lock for dispose of message
}

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
	u.conn.SetConn(skt, cancelSocket)
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

	// log.Trace("send message ->")
	if u.sendMessageReplys() == false {
		u.sendMessageRecords()
	}
}

// 发送消息回执
func (u *User) sendMessageReplys() bool {
	conn := u.conn
	if conn == nil {
		return false
	}

	first, tail := u.replys.Head()
	if first == nil { // there's no reply
		return false
	}

	err := conn.Send(first)
	if err == nil {
		u.replys = tail
		log.TraceF("user %s reply(%d left) -> %s", u.GetID(), len(u.replys), string(first))
		return true
	}
	return false
}

//  发送普通消息, 如果有数据发送, 返回 true
func (u *User) sendMessageRecords() bool {
	conn := u.conn
	if conn == nil {
		return false
	}

	if len(u.MessageRecords) <= 0 {
		return false
	}

	msgID := u.MessageRecords[0]
	msg := u.messageStore.GetMessage(msgID)
	if msg != nil {
		conn.Send(msg.GetContent())
		return true
	}
	// 该消息已经本体已经不存在, 发送记录中不应再保存
	go u.RemoveRecordCache(msgID)
	return false
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

// Release release resoures of user
func (u *User) Release() {
	u.lockConn.Lock()
	defer u.lockConn.Unlock()
	if u.conn != nil {
		u.conn.Close("")
	}
}

// NewMessage new data upload from user's conn
func (u *User) NewMessage(mIn *communication.Message) (err error) {
	switch mIn.Protocal {
	case communication.ProtoShare, communication.ProtoText:
		err = u.handleNormalMessage(mIn)
	case communication.ProtoReply:
		log.TraceF("user %s(%s) receive msg %s's reply", u.detail.GetID(), u.detail.GetName(), mIn.MessageID)
		u.RemoveRecordCache(mIn.GetID())
	default:
	}
	return
}

func (u *User) handleNormalMessage(mIn *communication.Message) error {
	err := mIn.Fill(u.detail.GetID(), u.detail.GetName())
	if err != nil {
		log.InfoF("read message error: %s", err)
		return err
	}
	log.TraceF("user %s(%s) receive new msg: %s", u.detail.GetID(), u.detail.GetName(), string(mIn.MessageBytes))
	go u.messageStore.PopNewMessage(mIn)

	replyMsg, e := communication.NewReplyMessage(mIn.MessageID)
	if e != nil {
		log.InfoF("NewReplyMessage error: %s", e)
		return e
	}
	u.replys = u.replys.append(replyMsg.MessageBytes)
	log.TraceF("user %s(%s) generate new reply: %s", u.detail.GetID(), u.detail.GetName(), string(replyMsg.MessageBytes))
	return nil
}

// NewDataIn new data upload from user's conn
func (u *User) NewDataIn(data []byte) error {
	var mIn communication.Message
	err := json.Unmarshal(data, &mIn)
	if err == nil {
		return u.NewMessage(&mIn)
	}
	log.SysF("NewMessage error: %s %s", err.Error(), string(data))
	return nil
}

//  If this msg is from this user, it should not be sent to this user's client again
func (u *User) isMsgMine(msg *communication.Message) bool {
	return u.GetID() == msg.ID
}

// AddMessageToCache stores msg id to cache, then msges will be sent to client
func (u *User) AddMessageToCache(message *communication.Message) {
	if !u.isMsgMine(message) {
		return
	}
	//如果为虚拟用户,那么在线时才会接收消息
	if u.detail.IsFake() && !u.online() {
		return
	}

	msgID := message.GetID()

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

func (u *User) online() bool {
	return u.conn.Online()
}
