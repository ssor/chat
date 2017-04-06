package fakeuser

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
	"xsbPro/chat/client/protocol"
	"xsbPro/chat/node/server/communication"

	"github.com/parnurzeal/gorequest"
)

var (
	dispatcherHost = "127.0.0.1:8092"
	id             = "iamafakeuser1"
	group          = "1466220644_1466220670378098957"
)

func requestNodeHost(t *testing.T) string {
	_, res, errs := gorequest.New().Get(protocol.FormatLoginURL(dispatcherHost, id, group)).End()
	if errs != nil {
		t.Fatal(errs)
	}
	// t.Log(res)
	var resMessage struct {
		Code int `json:"code"`
		Data struct {
			Hosts []string `json:"hosts"`
		} `json:"data"`
	}
	err := json.Unmarshal([]byte(res), &resMessage)
	if err != nil {
		t.Fatal(err)
	}
	if resMessage.Code != 0 {
		t.Fatal("no dispatched node")
	}
	if resMessage.Data.Hosts == nil || len(resMessage.Data.Hosts) <= 0 {
		t.Fatal("no node: ", res)
	}

	return resMessage.Data.Hosts[0]
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

// func TestReply(t *testing.T) {
// 	host := requestNodeHost(t)
// 	url := protocol.FormatConnectURL(host, id, group)
// 	fu := NewFakeUser(id, url)

// 	msg := createMsg(t, "abc")
// 	err := fu.SendMsg(msg.MessageBytes)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	time.Sleep(3 * time.Second)
// }

func createMsg(t *testing.T, content string) *communication.Message {

	now := time.Now()
	msgID := fmt.Sprintf("%s%d", id, now.Nanosecond())
	msg, err := communication.NewTextMessage(msgID, id, "")
	if err != nil {
		t.Fatal(err)
	}
	return msg
}
