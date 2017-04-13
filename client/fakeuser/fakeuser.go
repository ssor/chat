package fakeuser

import (
	"context"
	"errors"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ssor/chat/node/server/communication"
	"github.com/ssor/chat/node/server/connection"
	"github.com/ssor/log"
)

var (
	// ErrSendMsgBufferFull when send buffer is full
	ErrSendMsgBufferFull = errors.New("SendMsgBufferFull")
)

// FakeUser is a client, it will connect to node and chat in group
type FakeUser struct {
	id               string
	url              string
	socketReadWriter *connection.SocketReadWriter
	cancelReadBuffer context.CancelFunc
	sendMsgBuffer    chan []byte // msg tunnel for msg from upstream to sent to client
	messageInBuffer  chan *communication.Message
}

// NewFakeUser init a FakeUser, it connects to server
func NewFakeUser(id, url string) *FakeUser {
	fu := &FakeUser{
		id:              id,
		url:             url,
		sendMsgBuffer:   make(chan []byte, 1),
		messageInBuffer: make(chan *communication.Message, 64),
	}

	err := fu.run()
	if err != nil {
		log.SysF("err: %s", err.Error())
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
	go fu.receiveDataLoop(ctx)

	return nil
}

func (fu *FakeUser) receiveDataLoop(ctx context.Context) {
	for {
		select {
		case data := <-fu.socketReadWriter.NewData():
			log.TraceF(" <- %s : %s", fu.id, string(data))
			msg, err := communication.UnmarshalMessage(data)
			if err != nil {
				log.TraceF("data in format err: %s", err)
			} else {
				fu.cacheMessage(msg)
				fu.replyBack(msg)
			}
		case e := <-fu.socketReadWriter.Err():
			if e == connection.ErrSockerError {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (fu *FakeUser) cacheMessage(msg *communication.Message) {
	fu.messageInBuffer <- msg
}

func (fu *FakeUser) replyBack(msg *communication.Message) {
	switch msg.Protocal {
	case communication.ProtoText, communication.ProtoShare:
		//do reply
	}
}

// data will be cache in writer, but writer does not make sure the data will be sent to client eventually
func (fu *FakeUser) sendDataUnsafe(data []byte) {
	writer := fu.socketReadWriter
	if writer != nil {
		writer.SendDataUnsafe(data)
	}
}

func (fu *FakeUser) dialServer() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(fu.url, nil)
	if err != nil {
		log.InfoF("user %s -> url: %s", fu.id, fu.url)
		return nil, err
	}
	return conn, nil
}

// Data output data from client
func (fu *FakeUser) Data() <-chan *communication.Message {
	return fu.messageInBuffer
}

// SendMsg send data to client, if buffer is full, returns error
func (fu *FakeUser) SendMsg(data []byte) error {
	select {
	case fu.sendMsgBuffer <- data:
	default:
		return ErrSendMsgBufferFull
	}

	return nil
}

// Release release resouces of FakeUser and SocketReadWriter
func (fu *FakeUser) Release() {
	if fu.cancelReadBuffer != nil {
		fu.cancelReadBuffer()
	}
	if fu.socketReadWriter != nil {
		fu.socketReadWriter.Release()
	}
}
