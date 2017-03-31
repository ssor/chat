package connection

import (
	"context"
	"errors"
	"sync"
	"time"

	"xsbPro/chat/node/server/communication"
	"xsbPro/log"

	"github.com/gorilla/websocket"
)

//
// connection 代表终端与服务端的真正连接,能够读取终端发送的数据,能够向终端发送数据
// 由于具体的网络发送需要时间,因此其中有一个消息的缓存 send ,放置于缓存中的消息会被终端逐一发送至终端
//
//

var (
	errSocketSendFailed = errors.New("socket error")
	errBufferFull       = errors.New("buffer full")
	// ErrSocketReadError means socket read over
	ErrSocketReadError = errors.New(EventSocketReadError)
	// EventSocketReadError means socket read over
	EventSocketReadError = "socket read error"
	// EventSocketWriteError means socket write over
	EventSocketWriteError = "socket write error"
)

// Connection is an middleman between the websocket connection and the user.
// it buffers some data
// if err, notify dataStore, and close self
type Connection struct {
	uid string

	ws              Socket
	cancelReadPump  context.CancelFunc
	cancelWritePump context.CancelFunc
	lastActiveTime  time.Time

	sendMsgBuffer    chan []byte // Buffered channel of outbound messages.
	comesInMsgBuffer chan []byte // Buffered channel of outbound messages.

	peacefullyClosed bool
	socketLock       *sync.Mutex //in case of setting socket sync error
	dataStore        dataStore   // if data comes in, put it to store
	eventCallbacks   []func(string)

	pingPeriod    time.Duration
	writeDuration time.Duration
}

// NewConnection init a Connection
func NewConnection(uid string, dataStore dataStore, pingPeriod time.Duration, writeDuration time.Duration) *Connection {
	return &Connection{
		pingPeriod:       pingPeriod,
		writeDuration:    writeDuration,
		sendMsgBuffer:    make(chan []byte, 1),
		comesInMsgBuffer: make(chan []byte, 1024),
		uid:              uid,
		socketLock:       &sync.Mutex{},
		dataStore:        dataStore,
	}
}

// Run start goroutine for  writing msg from buffer to socket
func (c *Connection) Run() {
}

func (c *Connection) startReadWritePump(skt Socket) {
	cancelRead := c.cancelReadPump
	if cancelRead != nil {
		cancelRead()
		c.cancelReadPump = nil
	}

	cancelWrite := c.cancelWritePump
	if cancelWrite != nil {
		cancelWrite()
		c.cancelWritePump = nil
	}

	go func() {
		c.cancelWritePump = c.writePump(skt)
	}()
	go func() {
		c.cancelReadPump = c.readPump(skt)
	}()
}

// SetConn set a real connection for this user
// the real connection need reconnection usually
func (c *Connection) SetConn(conn Socket) {
	c.socketLock.Lock()
	defer c.socketLock.Unlock()

	skt := c.ws
	if skt != nil {
		log.InfoF("kick off user %s ", c.uid)
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, communication.ProtoCloseLoginOnOtherDevice)
		writeSocket(skt, websocket.CloseMessage, closeMsg, c.writeDuration)
	}

	c.ws = conn
	log.InfoF("user %s get online ", c.uid)
	c.lastActiveTime = time.Now() //记录活动时间
	c.startReadWritePump(conn)
}

// GetID return conn's unique id
func (c *Connection) GetID() string {
	return c.uid
}

// Send send binary data to client
func (c *Connection) Send(m []byte) error {
	// c.mutex.Lock()
	// defer c.mutex.Unlock()

	buffer := c.sendMsgBuffer
	if buffer != nil {
		select {
		case buffer <- m:
			return nil
		default:
			// close(c.send)
			// return errSocketSendFailed
			return errBufferFull
		}
	}
	return errors.New("NoBuffer")
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Connection) readPump(skt Socket) (cancel context.CancelFunc) {
	defer func() {
		c.invokeCallbacks(EventSocketReadError)
	}()

	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel() // cancel when we are finished

	errCh := lanchSocketListening(ctx, skt, c.comesInMsgBuffer)
	for {
		select {
		case data := <-c.comesInMsgBuffer:
			c.dataStore.PopNewData(data, nil)
		case <-errCh:
			c.dataStore.PopNewData(nil, ErrSocketReadError)
			return
		case <-ctx.Done():
			return
		}
	}
}

func lanchSocketListening(ctx context.Context, ws Socket, dataInBuffer chan []byte) chan error {
	temp := make(chan error, 1)

	go func() {
		defer close(temp)
		temp <- startListeningLoop(ctx, dataInBuffer, ws)
	}()
	return temp
}

func startListeningLoop(ctx context.Context, tunnel chan []byte, ws Socket) error {
	for {
		log.InfoF("start read message ...")
		startRead := time.Now().Unix()
		_, data, err := ws.ReadMessage()
		if err != nil {
			endRead := time.Now().Unix()
			log.InfoF("After %d seconds, read Message over", endRead-startRead)
			return err
		}
		select {
		case tunnel <- data:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// WritePump pumps messages from the buffer to the websocket connection.
func (c *Connection) writePump(skt Socket) (cancel context.CancelFunc) {
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		// c.ws = nil // ws error, should be set to nil
		ticker.Stop()
		c.invokeCallbacks(EventSocketWriteError)
		writeSocket(skt, websocket.CloseMessage, []byte{}, c.writeDuration)
		skt.Close()
	}()
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	for {
		select {
		case message, ok := <-c.sendMsgBuffer:
			if !ok { //服务端主动关闭
				writeSocket(skt, websocket.CloseMessage, []byte{}, c.writeDuration)
				return
			}
			if err := writeSocket(skt, websocket.TextMessage, message, c.writeDuration); err != nil {
				return
			}
		case <-ticker.C:
			if err := writeSocket(skt, websocket.PingMessage, []byte{}, c.writeDuration); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// Close close conn, write and read goroutine should return
func (c *Connection) Close(msg string, writeDuration time.Duration) {
	if c.cancelReadPump != nil {
		c.cancelReadPump()
		c.cancelReadPump = nil
	}
	if c.cancelWritePump != nil {
		c.cancelWritePump()
		c.cancelWritePump = nil
	}

	// close(c.sendMsgBuffer)
	// close(c.comesInMsgBuffer)

	// skt := c.ws
	// if skt == nil {
	// 	return
	// }
	// c.ws = nil
	// c.mutex.Lock()
	// defer c.mutex.Unlock()

	// c.peacefullyClosed = true //state that we close the conn because we want
	// if len(msg) > 0 {
	// 	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg)
	// 	writeSocket(skt, websocket.CloseMessage, closeMsg, writeDuration) //服务端主动关闭
	// } else {
	// 	writeSocket(skt, websocket.CloseMessage, []byte{}, writeDuration) //服务端主动关闭
	// }
	// skt.Close()
}

// AddEventCallback will add a handler for connection event
func (c *Connection) AddEventCallback(callback func(string)) {
	if c.eventCallbacks == nil {
		c.eventCallbacks = []func(string){}
	}
	c.eventCallbacks = append(c.eventCallbacks, callback)
}

// write writes a message with the given message type and payload.
func writeSocket(skt Socket, mt int, payload []byte, writeDuration time.Duration) error {
	if skt == nil {
		return nil
	}
	skt.SetWriteDeadline(time.Now().Add(writeDuration))
	return skt.WriteMessage(mt, payload)
	// return errors.New("no phisical connection")
}

// write writes a message with the given message type and payload.
func (c *Connection) invokeCallbacks(evt string) {
	if c.eventCallbacks == nil {
		return
	}
	for _, callback := range c.eventCallbacks {
		go callback(evt)
	}
}
