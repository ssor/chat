package server

import (
	"encoding/json"
	"sync"
	"time"
	"xsbPro/log"
	modeldb "xsbPro/xsbdb"
)

type FakeUserInfo struct {
	id string
}

func NewFakeUserInfo(id string) *FakeUserInfo {
	return &FakeUserInfo{id: id}
}

func (fui *FakeUserInfo) GetUserID() string {
	return fui.id
}

func (fui *FakeUserInfo) GetUserName() string {
	return "FakeUser"
}

func (rui *FakeUserInfo) IsFake() bool {
	return true
}

type RealUserInfo struct {
	*modeldb.User
}

func NewRealUserInfo(u *modeldb.User) *RealUserInfo {
	return &RealUserInfo{u}
}

func (rui *RealUserInfo) GetUserID() string {
	return rui.ID
}

func (rui *RealUserInfo) GetUserName() string {
	return rui.Name
}

func (rui *RealUserInfo) IsFake() bool {
	return false
}

type UserInfo interface {
	GetUserID() string
	GetUserName() string
	IsFake() bool
}

func NewUser(user UserInfo, hub *Hub) *User {
	return &User{
		User:           user,
		hub:            hub,
		MessageRecords: MessageRecordArray{},
		lastActiveTime: time.Now(),
		mutex:          &sync.Mutex{},
	}
}

type IConnection interface {
	Close(string, time.Duration)
	Send(*messageRecord) error
	GetID() string
}

//User in chat
type User struct {
	// *UserInfo
	User           UserInfo
	hub            *Hub
	conn           IConnection
	MessageRecords MessageRecordArray
	lastActiveTime time.Time
	mutex          *sync.Mutex
}

func (u *User) onLine() bool {
	return u.User.IsFake() == false && u.conn != nil
}

func (user *User) broadcastMessage() {
	conn := user.conn
	if conn == nil {
		return
	}

	if len(user.MessageRecords) <= 0 {
		return
	}

	user.clearMessageRecords()

	for _, mr := range user.MessageRecords {
		// if mr.SendState == msg_state_default {
		err := conn.Send(mr)
		if err != nil {
			// log.InfoF("broadcastMessage error: %s", err)
			// send data next time, records after this will not use this connection
			// return
			break
		}
		// }
	}
}

func (user *User) clearMessageRecords() {
	if len(user.MessageRecords) <= 0 {
		return
	}
	//如果第一条消息是回执消息,说明排名第二的消息已经被收到,应该清除前两条消息
	if user.MessageRecords[0].Protocal == protoReply {
		if len(user.MessageRecords) < 2 { //有回执,但没消息说明有错误发生
			user.MessageRecords = MessageRecordArray{}
			return
		}

		user.MessageRecords = user.MessageRecords[2:]
	}

	// nowSeconds := time.Now().Unix()
	secondsOf3DaysAgo := time.Now().AddDate(0, 0, -3).UnixNano()

	list := MessageRecordArray{}
	for _, r := range user.MessageRecords {
		// if r.Timestamp > secondsOf3DaysAgo && r.SendState != msg_state_sent {
		if r.Timestamp > secondsOf3DaysAgo {
			list = append(list, r)
		}

		// if len(r.Receivers) > 0 {
		// 	list = append(list, r)
		// }
	}
	if len(list) != len(user.MessageRecords) {
		log.TraceF("user %s  %d records to send", user.User.GetUserName(), len(list))
	}
	user.MessageRecords = list
}

//new conn set
func (u *User) SetConn(conn IConnection) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	c := u.conn
	if c != nil {
		log.InfoF("close user %s(%s) pre conn", u.User.GetUserID(), u.User.GetUserName())
		c.Close(protoCloseLoginOnOtherDevice, 3*time.Second)
		// u.setBufferedMsgToDefault()
	} else {
		log.InfoF("user %s(%s) not on line before", u.User.GetUserID(), u.User.GetUserName())
	}

	u.conn = conn
	u.lastActiveTime = time.Now() //记录活动时间
	// m, _ := newLoginMessage(u.User)
	// u.hub.newMessage(m, nil)
}

//conn error occors
//bufferred msg should be reset to default state
// uid used to identify whether it is the same conn
func (u *User) ConnError(uid string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	c := u.conn
	if c != nil {
		log.InfoF("close user %s(%s) conn for error", u.User.GetUserID(), u.User.GetUserName())
		//check if it is the same conn
		if c.GetID() != uid {
			c.Close(protoCloseLoginOnOtherDevice, 3*time.Second)
		} else {
			c.Close("", 3*time.Second)
		}
		// u.setBufferedMsgToDefault()
	}
	u.conn = nil

	//如果为虚拟用户,离线后清除消息
	if u.User.IsFake() {
		u.MessageRecords = u.MessageRecords[0:0]
	}

	// u.setBufferedMsgToDefault()
	// m, _ := newLogoutMessage(u.User)
	// u.hub.newMessage(m, nil)
}

// func (u *User) setBufferedMsgToDefault() {
// 	for _, r := range u.MessageRecords {
// 		if r.SendState == msg_state_buffering {
// 			r.SendState = msg_state_default
// 		}
// 	}
// }

// NewMessage new data upload from user's conn
func (u *User) NewMessage(data []byte) {
	var mIn Message
	err := json.Unmarshal(data, &mIn)
	// if err == nil && (mIn.Protocal == protoShare || mIn.Protocal == protoText) {
	if err == nil {
		switch mIn.Protocal {
		case protoShare, protoText:
			err = (&mIn).fill(u.User.GetUserID(), u.User.GetUserName())
			if err != nil {
				log.InfoF("read message error: %s", err)
				return
			}
			err = u.hub.NewMessage(&mIn, nil)
			if err != nil {
				log.SysF("NewMessage error: %s", err.Error())
			}
		case protoReply:
			err = u.hub.NewMessage(&mIn, []string{u.User.GetUserID()})
			if err != nil {
				log.SysF("NewMessage error: %s", err.Error())
			}
		default:
		}
		// (mIn.Protocal == protoShare || mIn.Protocal == protoText)
		// m, err := newMessage(mIn.Protocal, u.ID, u.Name, mIn.Content)
	} else {
		log.SysF("NewMessage error: %s %s", err.Error(), string(data))
	}
}

func (u *User) AddRecord(m *Message) {
	//如果为虚拟用户,那么不在线时不接收消息
	if u.User.IsFake() && u.conn == nil {
		return
	}

	r := NewMessageRecord(m)
	switch m.Protocal {
	// case protoLogout: //, protoLogin
	// u.MessageRecords = append(MessageRecordArray{r}, u.MessageRecords...)
	case protoReply:
		u.MessageRecords = append(MessageRecordArray{r}, u.MessageRecords...)
	default:
		u.MessageRecords = append(u.MessageRecords, r)
	}

	if len(u.MessageRecords) > maxMessageCountCache {
		u.MessageRecords = u.MessageRecords[:maxMessageCountCache]
	}
}

// // UserList use
// type UserList map[string]*User

// type UserArray []*User

// func (ul UserList) FilterByGroup(group string) UserList {
// 	list := make(UserList)

// 	for id, u := range ul {
// 		if u.Group == group {
// 			list[id] = u
// 		}
// 	}

// 	return list
// }

// // GetIDMap use
// func (ul UserList) GetIDMap() map[string]bool {
// 	msu := make(map[string]bool)
// 	for _, ui := range ul {
// 		msu[ui.ID] = true
// 	}
// 	return msu
// }
