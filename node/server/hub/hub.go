package hub

import (
	"errors"
	"time"

	"github.com/ssor/chat/node/server/communication"
	user "github.com/ssor/chat/node/server/user"
)

var (
	maxMessageCountCache = 1000
	//ErrUserAlreadyOnline raise an error
	ErrUserAlreadyOnline = errors.New("user already on line")
)

// Hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	group            string //user's group
	GroupUsers       UserList
	stopCh           chan bool
	eventSendMessage chan int     //开始发送消息的轮询
	messageCache     *messageList // stores all messages
}

func newHub(group string, users UserList) *Hub {
	h := Hub{
		group:            group,
		stopCh:           make(chan bool),
		eventSendMessage: make(chan int, 1),
		GroupUsers:       users,
		messageCache:     newMessageList(),
	}

	return &h
}

func (h *Hub) close() {
	select {
	case h.stopCh <- true:
	default:
	}
	for _, user := range h.GroupUsers {
		user.Release()
	}
}

// PopNewMessage use to receive msg for user
func (h *Hub) PopNewMessage(msg *communication.Message) {
	h.messageCache.add(msg)

	for _, u := range h.GroupUsers {
		u.AddMessageToCache(msg)
	}
}

// GetID return a uid for hub
func (h *Hub) GetID() string {
	return h.group
}

// GetMessage returns  content of message with id
func (h *Hub) GetMessage(id string) *communication.Message {
	return h.messageCache.find(id)
}

func (h *Hub) run() {
	tickerHour := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-h.eventSendMessage:
			// log.TraceF("notifyUserSendMessage ... ")
			h.notifyUserSendMessage()
		case <-tickerHour.C:
		case <-h.stopCh:
			return
		}
	}
}

func (h *Hub) sendMessge() {
	select {
	case h.eventSendMessage <- 1:
	default:
	}
}

func (h *Hub) notifyUserSendMessage() {
	for _, u := range h.GroupUsers {
		go u.SendMessage()
	}
}

// AddUser add user to group
func (h *Hub) AddUser(users ...*user.User) error {
	for _, u := range users {
		if h.FindUser(u.GetID()) == nil {
			h.GroupUsers = h.GroupUsers.add(u)
		} else {
			return ErrUserAlreadyOnline
		}
	}
	return nil
}

// RemoveUser removes user not existed
func (h *Hub) RemoveUser(u *user.User) error {
	h.GroupUsers, _ = h.GroupUsers.remove(u.GetID())
	return nil
}

// RefreshUsers will add new user and remove not existed user
func (h *Hub) RefreshUsers(users UserList) error {
	//new user added to this group
	for _, dbUser := range users {
		h.AddUser(dbUser)
	}
	for _, user := range h.GroupUsers {
		if users.find(user.GetID()) == nil {
			h.RemoveUser(user)
		}
	}

	return nil
}

// FindUser user a user in this hub
func (h *Hub) FindUser(id string) *user.User {
	return h.GroupUsers.find(id)
}
