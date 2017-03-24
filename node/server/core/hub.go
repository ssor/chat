package core

import (
	"errors"
	"time"
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
		eventSendMessage: make(chan int, 1024),
		GroupUsers:       users,
		messageCache:     newMessageList(),
	}
	for _, u := range h.GroupUsers {
		u.SetMessagePopHandler(h.newMessage)
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

func (h *Hub) newMessage(m message) error {
	h.messageCache.add(m)

	for _, u := range h.GroupUsers {
		u.AddMessageToCache(m)
	}

	return nil
}

func (h *Hub) run() {
	tickerHour := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-h.eventSendMessage:
			for _, u := range h.GroupUsers {
				go u.BroadcastMessage()
			}
		case <-tickerHour.C:
		case <-h.stopCh:
			return
		}
	}
}

// AddUser add user to group
func (h *Hub) AddUser(u user) error {
	if h.findUser(u.GetID()) == nil {
		h.GroupUsers = h.GroupUsers.add(u)
	} else {
		return ErrUserAlreadyOnline
	}
	return nil
}

// RemoveUser removes user not existed
func (h *Hub) RemoveUser(u user) error {
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

func (h *Hub) findUser(id string) user {
	return h.GroupUsers.find(id)
}
