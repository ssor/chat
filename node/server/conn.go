package server

import (
	"errors"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	// "net/http"
	"time"
	"xsbPro/log"
)

//
// connection 代表终端与服务端的真正连接,能够读取终端发送的数据,能够向终端发送数据
// 由于具体的网络发送需要时间,因此其中有一个消息的缓存 send ,放置于缓存中的消息会被终端逐一发送至终端
//
//

var (
	errSocketSendFailed = errors.New("socket error")
	err_buffer_full     = errors.New("buffer full")
)

type ReportObject interface {
	NewMessage([]byte)
	ConnError(string)
}

type Socket interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	SetWriteDeadline(t time.Time) error
	Close() error
}

// Connection is an middleman between the websocket connection and the hub.
type Connection struct {
	// The websocket connection.
	// ws *websocket.Conn
	ws               Socket
	reportObj        ReportObject
	send_msg_buffer  chan *messageRecord // Buffered channel of outbound messages.
	uid              string
	peacefullyClosed bool
	mutex            *sync.Mutex //防止 send_msg_buffer 关闭后让发送信息
}

func NewConnection(ws Socket, reportObj ReportObject, uid string) *Connection {
	return &Connection{
		ws:              ws,
		send_msg_buffer: make(chan *messageRecord),
		reportObj:       reportObj,
		uid:             uid + "_" + strconv.FormatInt(time.Now().UnixNano(), 16),
		mutex:           &sync.Mutex{},
	}
}

func (c *Connection) GetID() string {
	return c.uid
}

func (c *Connection) Send(m *messageRecord) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.send_msg_buffer != nil {
		select {
		case c.send_msg_buffer <- m:
		default:
			// close(c.send)
			// return errSocketSendFailed
			return err_buffer_full
		}
	}
	return nil
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Connection) ReadPump() {
	defer func() {
		// c.Hub.unregisterConn(c)
		if c.peacefullyClosed == false && c.reportObj != nil {
			c.reportObj.ConnError(c.uid)
		}
	}()
	// c.ws.SetReadLimit(maxMessageSize)
	// c.ws.SetReadDeadline(time.Now().Add(pongWait))
	// c.ws.SetPongHandler(func(string) error {
	// 	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })

	for {
		log.InfoF("start read message ...")
		startRead := time.Now().Unix()
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			endRead := time.Now().Unix()
			log.InfoF("After %d seconds, read Message over", endRead-startRead)
			break
		}

		c.reportObj.NewMessage(data)
	}
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte, writeDuration time.Duration) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeDuration))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Connection) WritePump(pingPeriod time.Duration, writeDuration time.Duration) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// c.Hub.unregisterConn(c)
		if c.peacefullyClosed == false && c.reportObj != nil {
			c.reportObj.ConnError(c.uid)
		}
	}()

	for {
		select {
		case message, ok := <-c.send_msg_buffer:
			if !ok { //服务端主动关闭
				c.peacefullyClosed = true
				c.write(websocket.CloseMessage, []byte{}, writeDuration)
				return
			}
			if err := c.write(websocket.TextMessage, message.MessageBytes, writeDuration); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}, writeDuration); err != nil {
				return
			}
		}
	}
}

func (c *Connection) Close(msg string, writeDuration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.peacefullyClosed = true //state that we close the conn because we want
	if len(msg) > 0 {
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg)
		c.write(websocket.CloseMessage, closeMsg, writeDuration) //服务端主动关闭
	} else {
		c.write(websocket.CloseMessage, []byte{}, writeDuration) //服务端主动关闭
	}
	c.ws.Close()
	// c.ws = nil
	if c.send_msg_buffer != nil {
		select {
		case _, open := <-c.send_msg_buffer:
			if open {
				close(c.send_msg_buffer)
			}
		default:
			close(c.send_msg_buffer)
		}

		c.send_msg_buffer = nil
	}
}
