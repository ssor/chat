package group

import (
	"testing"
	"time"

	"github.com/ssor/chat/client/protocol"
)

var (
	dispatcherHost = "127.0.0.1:8092"
	id             = "iamafakeuser1"
	group          = "1466220644_1466220670378098957"
)

func TestInitGroup(t *testing.T) {

	nodeHost, err := protocol.RequestNodeHost(dispatcherHost, group)
	if err != nil {
		t.Fatal(err)
	}

	group := NewGroup(3, group, nodeHost)

	// time.Sleep(1 * time.Second)
	// msg := createMsg(t, "test")
	group.SendMessage("test")

	time.Sleep(10 * time.Second)

	group.DumpMessage()
}

// func createMsg(t *testing.T, content string) *communication.Message {

// 	now := time.Now()
// 	msgID := fmt.Sprintf("%s%d", id, now.Nanosecond())
// 	msg, err := communication.NewTextMessage(msgID, id, content)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	return msg
// }
