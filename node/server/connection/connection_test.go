package connection

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

var (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// pingPeriod = 5 * time.Second
)

func TestNormalSocket(t *testing.T) {
	fs := newFakeSocket()

	msgList := [](string){"a", "b", "c"}

	dr := &dataReceivor{
		msgToCheck: msgList,
	}

	waitForReadOver := make(chan bool)
	conn := NewConnection("1", dr, pingPeriod, writeWait)
	conn.AddEventCallback(func(evt string) {
		if evt == EventSocketReadError {
			close(waitForReadOver)
		}
	})
	conn.SetConn(fs)
	// conn.Run(pingPeriod, writeWait)

	msgTunnel := make(chan string, len(msgList)) // simulate a msg source, until msg send ok, will not removed
	for _, s := range msgList {
		msgTunnel <- s
	}

	sendMsgFromTunnel(msgTunnel, conn.Send)

	<-waitForReadOver
	if len(dr.msgToCheck) > 0 {
		spew.Dump(dr.msgToCheck)
		t.Fatal("not all msg received")
	}

}

func TestSocketWriteTimeout(t *testing.T) {

}

func sendMsgFromTunnel(tunnel chan string, send func([]byte) error) {
	for {
		select {
		case s := <-tunnel:
			err := send([]byte(s))
			if err != nil {
				fmt.Println(err, " -> ", s)
				// if send failed, restore and send again next time
				time.Sleep(10 * time.Millisecond)
				tunnel <- s
			} else {
				fmt.Println("send OK -> ", s)
			}
		default:
			return
		}
	}
}

type dataReceivor struct {
	msgToCheck []string
}

func (dr *dataReceivor) PopNewData(bs []byte, err error) {
	fmt.Println("pop up msg -> ", string(bs))
	list := []string{}
	for _, s := range dr.msgToCheck {
		if s != string(bs) {
			list = append(list, s)
		}
	}
	dr.msgToCheck = list
}

// fakeSocket when msg comes in, will be sent out
type fakeSocket struct {
	msgCache chan *socketMessage
}

type socketMessage struct {
	messageType int
	content     []byte
}

func newFakeSocket() *fakeSocket {
	fs := &fakeSocket{
		msgCache: make(chan *socketMessage, 5),
	}
	return fs
}

func (fs *fakeSocket) ReadMessage() (messageType int, p []byte, err error) {
	time.Sleep(100 * time.Millisecond)
	select {
	case msg := <-fs.msgCache:
		return msg.messageType, msg.content, nil
	default:
	}
	return 0, nil, errors.New("no message left")
}
func (fs *fakeSocket) WriteMessage(messageType int, data []byte) error {
	fs.msgCache <- &socketMessage{messageType, data}
	return nil
}
func (fs *fakeSocket) SetWriteDeadline(t time.Time) error {
	return nil
}
func (fs *fakeSocket) Close() error {
	return nil
}
