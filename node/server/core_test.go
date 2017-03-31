package server

import (
	"fmt"
	"testing"
	"time"
	"xsbPro/chat/node/server/user"
)

// coreUser
type coreUser struct {
	id           string
	ticker       *time.Ticker // simulate to say something
	stop         chan bool
	messageCache chan string
	messageStore common.MessageStore
}

func newCoreUser(id string, msgStore user.MessageStore) *coreUser {
	u := &coreUser{
		id:           id,
		ticker:       time.NewTicker(3 * time.Second),
		messageCache: make(chan string, 1024),
		messageStore: msgStore,
	}
	go u.run()
	return u
}

func (cu *coreUser) run() {
	for {
		select {
		case t := <-cu.ticker.C:
			if cu.messageStore != nil {
				fmt.Println(cu.id, " say at ", t.Format(time.RFC3339))
				id := fmt.Sprintf("userID_%s", cu.id)
				name := fmt.Sprintf("userName_%s", cu.id)
				msg, _ := connect.NewMessage(connect.ProtoText, id, name, t.Format(time.RFC1123Z))
				cu.messageStore.PopNewMessage(msg)
			} else {
				panic("cannot say")
			}
		case <-cu.stop:
			return
		}
	}
}

func (cu *coreUser) SendMessage() {
}

func (cu *coreUser) GetID() string {
	return cu.id
}
func (cu *coreUser) AddMessageToCache(msgID string) {
	if len(cu.messageCache) < 1024 {
		cu.messageCache <- msgID
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
			fmt.Println(cu.id, " ---> send message to client: ", msg)
		default:
			return
		}
	}
}

func (cu *coreUser) RemoveRecordCache(id string) {
}
func (cu *coreUser) Release() {
	close(cu.stop)
}

var ()

func init() {
}

func TestCore(t *testing.T) {
	hm := NewHubManager()
	hub := hm.AddHub("1", nil)
	users := []interface{}{}
	for index := 0; index < 3; index++ {
		users = append(users, newCoreUser(fmt.Sprintf("user_%d", index), hub))
	}
	hm.AddHub("1", ToUserList(users...))
	time.Sleep(100 * time.Second)
}
