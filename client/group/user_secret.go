package group

import (
	"context"
	"fmt"
	"time"

	"github.com/ssor/chat/client/fakeuser"
	"github.com/ssor/chat/client/protocol"
	"github.com/ssor/chat/node/server/communication"
	"github.com/ssor/log"
)

// UserSecret collect messages of fakeuser, and analysis them
type UserSecret struct {
	id                string
	user              *fakeuser.FakeUser
	messages          []*communication.Message
	cancelReceiveData context.CancelFunc
}

type UserSecretList []*UserSecret

func NewUserSecret(id, groupID, nodeHost string) *UserSecret {
	us := &UserSecret{
		id:       id,
		messages: []*communication.Message{},
	}

	url := protocol.FormatConnectURL(nodeHost, id, groupID)
	us.user = fakeuser.NewFakeUser(id, url)
	if us.user == nil {
		panic("create fake user error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	us.cancelReceiveData = cancel
	go us.receiveData(ctx)
	return us
}

func (us *UserSecret) sendMessage(content string) error {
	msg, err := us.createMsg(content)
	if err != nil {
		return err
	}
	log.TraceF("user %s content -> %s [%s]", us.id, content, string(msg.MessageBytes))
	return us.user.SendMsg(msg.MessageBytes)
}
func (us *UserSecret) receiveData(ctx context.Context) {
	if us.user == nil {
		panic("user nil")
	}

	for {
		select {
		case data := <-us.user.Data():
			msg, err := communication.UnmarshalMessage(data)
			if err != nil {
				log.SysF("cannot parse message: %s", err.Error())
			} else {
				us.messages = append(us.messages, msg)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (us *UserSecret) Release() {
	cancel := us.cancelReceiveData
	if cancel != nil {
		cancel()
	}
}

func (us *UserSecret) createMsg(content string) (*communication.Message, error) {
	now := time.Now()
	msgID := fmt.Sprintf("%s%d", us.id, now.Nanosecond())
	msg, err := communication.NewTextMessage(msgID, us.id, content)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
