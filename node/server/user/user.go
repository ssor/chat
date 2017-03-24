package user

import (
	"encoding/json"
	"sync"
	"time"
	"xsbPro/log"
	modeldb "xsbPro/xsbdb"
)

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

func NewUser(user UserInfo, hub *Hub) *User {
	return &User{
		User:           user,
		hub:            hub,
		MessageRecords: messageRecordArray{},
		lastActiveTime: time.Now(),
		mutex:          &sync.Mutex{},
	}
}

//User in chat
type User struct {
	// *UserInfo
	User           UserInfo
	hub            *Hub
	conn           IConnection
	MessageRecords messageRecordArray
	lastActiveTime time.Time
	mutex          *sync.Mutex
}

func (u *User) onLine() bool {
	return u.User.IsFake() == false && u.conn != nil
}

func (u *User) broadcastMessage() {
	conn := u.conn
	if conn == nil {
		return
	}

	if len(u.MessageRecords) <= 0 {
		return
	}

	u.clearMessageRecords()

	for _, mr := range u.MessageRecords {
		err := conn.Send(mr)
		if err != nil {
			break
		}
	}
}

func (u *User) clearMessageRecords() {
	if len(u.MessageRecords) <= 0 {
		return
	}
	//如果第一条消息是回执消息,说明排名第二的消息已经被收到,应该清除前两条消息
	if u.MessageRecords[0].Protocal == protoReply {
		if len(u.MessageRecords) < 2 { //有回执,但没消息说明有错误发生
			u.MessageRecords = messageRecordArray{}
			return
		}

		u.MessageRecords = u.MessageRecords[2:]
	}

	// nowSeconds := time.Now().Unix()
	secondsOf3DaysAgo := time.Now().AddDate(0, 0, -3).UnixNano()

	list := messageRecordArray{}
	for _, r := range u.MessageRecords {
		// if r.Timestamp > secondsOf3DaysAgo && r.SendState != msg_state_sent {
		if r.Timestamp > secondsOf3DaysAgo {
			list = append(list, r)
		}
	}
	if len(list) != len(u.MessageRecords) {
		log.TraceF("user %s  %d records to send", u.User.GetUserName(), len(list))
	}
	u.MessageRecords = list
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

}

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
	} else {
		log.SysF("NewMessage error: %s %s", err.Error(), string(data))
	}
}

func (u *User) AddRecord(m *Message) {
	//如果为虚拟用户,那么不在线时不接收消息
	if u.User.IsFake() && u.conn == nil {
		return
	}

	r := newMessageRecord(m)
	switch m.Protocal {
	case protoReply:
		u.MessageRecords = append(messageRecordArray{r}, u.MessageRecords...)
	default:
		u.MessageRecords = append(u.MessageRecords, r)
	}

	if len(u.MessageRecords) > maxMessageCountCache {
		u.MessageRecords = u.MessageRecords[:maxMessageCountCache]
	}
}
