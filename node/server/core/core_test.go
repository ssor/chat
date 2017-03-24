package core

import (
	"fmt"
	"testing"
	"time"
	"xsbPro/chat/node/server/protocol"
)

// coreUser
type coreUser struct {
	id             string
	messageHandler func(message) error
	ticker         *time.Ticker
	stop           chan bool
	messageCache   chan message
}

func newCoreUser(id string) *coreUser {
	u := &coreUser{
		id:           id,
		ticker:       time.NewTicker(3 * time.Second),
		messageCache: make(chan message, 1024),
	}
	go u.run()
	return u
}

func (cu *coreUser) run() {
	for {
		select {
		case t := <-cu.ticker.C:
			if cu.messageHandler != nil {
				fmt.Println(cu.id, " raise message at ", t.Format(time.RFC3339))
				id := fmt.Sprintf("userID_%s", cu.id)
				name := fmt.Sprintf("userName_%s", cu.id)
				msg, _ := protocol.NewMessage(protocol.ProtoText, id, name, t.Format(time.RFC1123Z))
				cu.messageHandler(msg)
			} else {
				panic("messageHandler is nil")
			}
		case <-cu.stop:
			return
		}
	}
}

func (cu *coreUser) GetID() string {
	return cu.id
}
func (cu *coreUser) AddMessageToCache(msg message) {
	if len(cu.messageCache) < 1024 {
		cu.messageCache <- msg
	} else {
		fmt.Println("cannot cache message")
	}
}
func (cu *coreUser) BroadcastMessage() {
	defer func() {
		fmt.Println("<--- quit BroadcastMessage, leaves messages: ", len(cu.messageCache))
	}()

	for {
		select {
		case msg := <-cu.messageCache:
			fmt.Println(cu.id, " ---> send message to client: ", string(msg.GetContent()))
		default:
			return
		}
	}
}
func (cu *coreUser) SetMessagePopHandler(f func(message) error) {
	cu.messageHandler = f
}
func (cu *coreUser) Release() {
	close(cu.stop)
}

var ()

func init() {
}

func TestCore(t *testing.T) {
	hm := NewHubManager()
	users := []interface{}{}
	for index := 0; index < 3; index++ {
		users = append(users, newCoreUser(fmt.Sprintf("user_%d", index)))
	}
	hm.AddHub("1", ToUserList(users...))
	time.Sleep(100 * time.Second)
}
