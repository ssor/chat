package fakeuser

import (
	"context"
	"errors"
	"xsbPro/chat/node/server/connection"
	"xsbPro/log"

	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrSendMsgBufferFull = errors.New("SendMsgBufferFull")
)

// FakeUser is a client, it will connect to node and chat in group
type FakeUser struct {
	id               string
	url              string
	socketReadWriter *connection.SocketReadWriter
	cancelReadBuffer context.CancelFunc
	sendMsgBuffer    chan []byte // msg tunnel for msg from upstream to sent to client
}

func NewFakeUser(id, url string) *FakeUser {
	fu := &FakeUser{
		id:            id,
		url:           url,
		sendMsgBuffer: make(chan []byte, 1),
	}

	err := fu.run()
	if err != nil {
		return nil
	}
	return fu
}

// start goroutines to handle msg
func (fu *FakeUser) run() error {
	ws, err := fu.dialServer()
	if err != nil {
		return err
	}
	srw := connection.NewSocketReadWriter(ws, fu.sendMsgBuffer, 5*time.Second)
	fu.socketReadWriter = srw
	ctx, cancel := context.WithCancel(context.Background())
	fu.cancelReadBuffer = cancel

	go func() {
		for {
			select {
			case data := <-srw.NewData():
				log.TraceF(" <- %s : %s", fu.id, string(data))
			case e := <-srw.Err():
				if e == connection.ErrSockerError {
					// fu.run()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (fu *FakeUser) dialServer() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(fu.url, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (fu *FakeUser) SendMsg(data []byte) error {
	// return fu.socketReadWriter.SendData(data)
	select {
	case fu.sendMsgBuffer <- data:
	default:
		return ErrSendMsgBufferFull
	}
	// ws, err := fu.dialServer()
	// if err != nil {
	// 	return err
	// }
	// err = ws.WriteMessage(websocket.TextMessage, data)
	// if err != nil {
	// 	log.SysF("SendMsg err: %s", err)
	// 	return err
	// }
	return nil
}

func (fu *FakeUser) Release() {
	if fu.cancelReadBuffer != nil {
		fu.cancelReadBuffer()
	}
	if fu.socketReadWriter != nil {
		fu.socketReadWriter.Release()
	}
}

// func (fu *FakeUser) ReceiveMsg() ([]byte, error) {
// 	ws, err := fu.dialServer()
// 	if err != nil {
// 		return nil, err
// 	}
// 	_, data, err := ws.ReadMessage()
// 	if err != nil {
// 		log.SysF("user %s read err: %s", fu.id, err)
// 		return nil, err
// 	}
// 	return data, nil
// }
