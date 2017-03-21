package server

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
	"xsbPro/chatDispatcher/lua"
	db "xsbPro/xsbdb"
)

var (
	maxMessageCountCache = 1000

	errUserAlreadyOnline = errors.New("user already on line")
)

type connectionRequest struct {
	// conn  *connection
	conn  IConnection
	chRes chan error
}

// HubList use
type HubList map[string]*Hub

// Hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	group string //user's group
	// Register requests from the connections.
	register chan *connectionRequest

	// Unregister requests from connections.
	unregister  chan *connectionRequest
	investigate chan *Questionnaire

	// MessageRecords MessageRecordArray

	recordCh chan []*messageRecord

	//同步发送消息状态;如果正在群发消息过长时,又有要求群发消息的事件发生,则首先检查该事件是否发生,如果正在发生,则略过此次触发
	broadcasting bool

	stopCh           chan bool
	eventSendMessage chan int //开始发送消息的轮询

	GroupUsers *SafeUserList
	// groupUsers UserList

	currentInstantMessageCount int64
	mutex                      *sync.Mutex

	*HubManager
}

func NewHub(group string, users *SafeUserList) *Hub {
	h := Hub{
		// broadcast:      make(chan *Message, 1024),
		group:       group,
		recordCh:    make(chan []*messageRecord, 1024),
		register:    make(chan *connectionRequest, 1024),
		unregister:  make(chan *connectionRequest, 1024),
		investigate: make(chan *Questionnaire),

		stopCh:           make(chan bool),
		eventSendMessage: make(chan int, 1024),
		// connections:      make(ConnectionList),
		// MessageRecords: MessageRecordArray{},
		mutex:      &sync.Mutex{},
		GroupUsers: users,
	}

	for _, u := range users.Items() {
		u.hub = &h
	}

	return &h
}

func (h *Hub) close() {
	select {
	case h.stopCh <- true:
	default:
	}
	for _, user := range h.GroupUsers.Items() {
		user.conn.Close("", 1*time.Second)
	}
}

func (h *Hub) RefreshUsers(scriptExecutor lua.ScriptExecutor) error {
	users, err := lua.GetGroupUsersFromCache(h.group, scriptExecutor)
	if err != nil {
		return err
	}

	//new user added to this group
	for _, db_user := range users {
		if h.FindUser(db_user.ID) == nil {
			nu := NewUser(NewRealUserInfo(db_user), h)
			h.GroupUsers.Set(db_user.ID, nu)
		}
	}
	//some user removed from this group
	find_user := func(users db.UserArray, id string) *db.User {
		for _, user := range users {
			if user.ID == id {
				return user
			}
		}
		return nil
	}
	for _, user := range h.GroupUsers.Items() {
		if find_user(users, user.User.GetUserID()) == nil {
			user.conn.Close("REMOVED_FROM_GROUP", 3*time.Second)
			h.GroupUsers.Delete(user.User.GetUserID())
		}
	}

	return nil
}

func (h *Hub) acceptInterview(questionnaire *Questionnaire) {
	h.investigate <- questionnaire
}

func contains_string(slice []string, s string) bool {
	if slice == nil {
		return false
	}

	for _, in_s := range slice {
		if in_s == s {
			return true
		}
	}

	return false

}

func (h *Hub) NewMessage(m *Message, destUsers []string) error {
	bs, err := json.Marshal(m)
	if err != nil {
		return err
	}
	m.MessageBytes = bs

	for _, u := range h.GroupUsers.Items() {
		if destUsers == nil { //为空时默认为全部
			u.AddRecord(m)
		} else {
			if contains_string(destUsers, u.User.GetUserID()) {
				u.AddRecord(m)
			}
		}
	}

	// // now := time.Now().Unix()
	// records := MessageRecordArray{}
	// if destUsers == nil {
	// 	// record.Receivers = h.getReceivers()
	// 	destUsers = h.getReceivers()

	// }
	// for _, userID := range destUsers {

	// 	record := &messageRecord{}
	// 	record.message = m
	// 	record.Receiver = userID
	// 	// record.CreateTime = now
	// 	records = append(records, record)
	// }

	// // log.TraceF("--> %d receivers got", len(record.Receivers))

	// bs, err := json.Marshal(m)
	// if err != nil {
	// 	return err
	// }
	// m.MessageBytes = bs
	// // record.MessageBytes = bs

	// h.recordCh <- records

	return nil
}

func (h *Hub) getReceivers() []string {
	receivers := []string{}
	// receivers := make(map[string]bool)
	for _, u := range h.GroupUsers.Items() {
		// receivers[u.ID] = true
		receivers = append(receivers, u.User.GetUserID())
	}
	return receivers
}

func (h *Hub) run() {
	ticker_hour := time.NewTicker(1 * time.Hour)
	for {
		select {
		case questionnare := <-h.investigate:
			onLineCount := 0
			for _, u := range h.GroupUsers.Items() {
				if u.onLine() == true {
					onLineCount++
				}
			}
			questionnare.RecycleChan <- NewInvestigationReport(h.group, h.GroupUsers.Length(), onLineCount, 0, 0)
		case <-h.eventSendMessage:
			if h.broadcasting == false {
				h.broadcasting = true
				go func() {
					// var done sync.WaitGroup
					// done.Add(h.groupUsers.Length())

					for _, u := range h.GroupUsers.Items() {
						u.broadcastMessage()
					}
					h.broadcasting = false
				}()
				// h.clearMessageRecords()
				// go h.broadcastMessage(h.connections.Filter(nil))
			}
		case <-ticker_hour.C:
			//hub 内部逻辑处理
			now := time.Now()
			//如果非活动状态持续时间超过72小时,则将其删除
			for _, u := range h.GroupUsers.Items() {
				if u.User.IsFake() && now.Sub(u.lastActiveTime) > time.Hour*72 {
					h.GroupUsers.Delete(u.User.GetUserID())
				}
			}
		case <-h.stopCh:
			return
		}
	}
}

func (h *Hub) FindUser(id string) *User {
	u := h.GroupUsers.Get(id)
	return u
}

// //TODO:需要优化,应该按照个人发送消息,而不应该每次轮询所有消息
// func (h *Hub) broadcastMessage(connections ConnectionList) {
// 	// log.TraceF("broadcastMessage ...")
// 	if len(h.MessageRecords) <= 0 {
// 		h.broadcasting = false
// 		return
// 	}
// 	for _, mr := range h.MessageRecords {
// 		// r.sendMessage(connections)

// 		c, exists := connections.Find(mr.Receiver)
// 		if exists == false { //use is offline
// 			return
// 		}

// 		err := c.Send(mr)
// 		if err != nil {
// 			log.InfoF("broadcastMessage error: %s", err)
// 			// send data next time, records after this will not use this connection
// 			return
// 		}
// 	}

// 	h.broadcasting = false
// }

// //有些 user 已经不在 hub 中,将其从记录中移除,防止记录无法删除

// //将已经发送完毕的记录删除
// //只保留 *三天(默认)* 内的聊天记录
// func (h *Hub) clearMessageRecords() {
// 	if len(h.MessageRecords) <= 0 {
// 		return
// 	}

// 	// nowSeconds := time.Now().Unix()
// 	secondsOf3DaysAgo := time.Now().AddDate(0, 0, -3).Unix()

// 	list := MessageRecordArray{}
// 	for _, r := range h.MessageRecords {
// 		if r.CreateTime > secondsOf3DaysAgo && r.SendState == false {
// 			list = append(list, r)
// 		}

// 		// for id := range r.Receivers {
// 		// 	if h.groupUsers.Check(id) == false {
// 		// 		log.TraceF("group %s user %s removed as not in group any more", h.group, id)
// 		// 		delete(r.Receivers, id)
// 		// 	}
// 		// }

// 		// if len(r.Receivers) > 0 {
// 		// 	list = append(list, r)
// 		// }
// 	}
// 	if len(list) != len(h.MessageRecords) {
// 		log.TraceF("group %s  %d records left", h.group, len(list))
// 	}
// 	h.MessageRecords = list
// }
