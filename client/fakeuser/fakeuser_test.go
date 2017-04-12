package fakeuser

import (
	"fmt"
	"testing"
	"time"

	"github.com/ssor/chat/client/protocol"
	"github.com/ssor/chat/node/server/communication"
)

var (
	dispatcherHost = "127.0.0.1:8092"
	id             = "iamafakeuser1"
	group          = "1466220644_1466220670378098957"
)

func requestNodeHost(t *testing.T) string {
	host, err := protocol.RequestNodeHost(dispatcherHost, group)
	if err != nil {
		t.Fatal(err)
	}
	return host
}

func TestSendMsg(t *testing.T) {
	host := requestNodeHost(t)
	url := protocol.FormatConnectURL(host, id, group)
	fu := NewFakeUser(id, url)

	msg := createMsg(t, "abc")

	err := fu.SendMsg(msg.MessageBytes)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(15 * time.Second)
}

func TestDuplicatedLogin(t *testing.T) {
	host := requestNodeHost(t)
	url := protocol.FormatConnectURL(host, id, group)
	NewFakeUser(id, url)

	time.Sleep(3 * time.Second)

	NewFakeUser(id, url)

	time.Sleep(3 * time.Second)
}

func createMsg(t *testing.T, content string) *communication.Message {

	now := time.Now()
	msgID := fmt.Sprintf("%s%d", id, now.Nanosecond())
	msg, err := communication.NewTextMessage(msgID, id, "")
	if err != nil {
		t.Fatal(err)
	}
	return msg
}
