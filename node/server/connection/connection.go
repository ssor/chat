package connection

import (
	"context"
	"errors"
	"sync"
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

	cancelLoop       context.CancelFunc
	socketReadWriter *SocketReadWriter

	lastActiveTime time.Time

	sendMsgBuffer chan []byte // Buffered channel of outbound messages.

	socketLock *sync.Mutex //in case of setting socket sync error
	dataStore  dataStore   // if data comes in, put it to store

	pingPeriod    time.Duration
	writeDuration time.Duration
}

// NewConnection init a Connection
func NewConnection(uid string, dataStore dataStore, pingPeriod time.Duration, writeDuration time.Duration) *Connection {
	return &Connection{
		uid:           uid,
		pingPeriod:    pingPeriod,
		writeDuration: writeDuration,
		sendMsgBuffer: make(chan []byte, 1),
		socketLock:    &sync.Mutex{},
		dataStore:     dataStore,
	}
}

// Online returns false if client connecting to server
func (c *Connection) Online() bool {
	after := time.After(500 * time.Millisecond)
	select {
	// case <-c.bridgeForSockets:
	case <-c.socketReadWriter.Online():
		return true
	case <-after:
		return false
	}
}

func (c *Connection) startReadWritePump(skt Socket, cancelSocket context.CancelFunc) {
	srw := NewSocketReadWriter(skt, c.sendMsgBuffer, 5*time.Second)
	c.socketReadWriter = srw
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelLoop = cancel

	go func() {
		defer func() {
			if cancelSocket != nil {
				cancelSocket()
			}
		}()
		for {
			select {
			case data := <-srw.NewData():
				log.TraceF(" <- %s : %s", c.GetID(), string(data))
				c.dataStore.NewDataIn(data)
			case e := <-srw.Err():
				if e == ErrSockerError {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// SetConn set a real connection for this user
// the real connection need reconnection usually
func (c *Connection) SetConn(conn Socket, cancelSocket context.CancelFunc) {
	c.socketLock.Lock()
	defer c.socketLock.Unlock()

	log.InfoF("user %s get online ", c.uid)
	c.lastActiveTime = time.Now() //记录活动时间

	// log.TraceF("Set Conn ---> ")
	c.Close("")
	c.startReadWritePump(conn, cancelSocket)
}

// GetID return conn's unique id
func (c *Connection) GetID() string {
	return c.uid
}

// Send send binary data to client
func (c *Connection) Send(m []byte) error {
	buffer := c.sendMsgBuffer
	if buffer != nil {
		select {
		case buffer <- m:
			return nil
		default:
			return errBufferFull
		}
	}
	return errors.New("NoBuffer")
}

// Close close conn, write and read goroutine should return
func (c *Connection) Close(msg string) {
	cancel := c.cancelLoop
	if cancel != nil {
		cancel()
	}
	srw := c.socketReadWriter
	if srw != nil {
		log.Trace("close -> release rw")
		srw.Release()
	} else {
		log.Trace("close -> rw is nil")
	}
}

// // ReadPump pumps messages from the websocket connection to the hub.
// func (c *Connection) readPump(skt Socket) (cancel context.CancelFunc) {
// 	defer func() {
// 		c.invokeCallbacks(EventSocketReadError)
// 		closeSocket(skt, false, c.uid)
// 	}()

// 	var ctx context.Context
// 	ctx, cancel = context.WithCancel(context.Background())
// 	defer cancel() // cancel when we are finished

// 	errCh := lanchSocketListening(ctx, skt, c.comesInMsgBuffer)
// 	for {
// 		select {
// 		case data := <-c.comesInMsgBuffer:
// 			c.dataStore.NewDataIn(data)
// 		case <-errCh:
// 			// c.dataStore.PopNewData(nil, ErrSocketReadError)
// 			return
// 		case <-ctx.Done():
// 			return
// 		}
// 	}
// }

// func lanchSocketListening(ctx context.Context, ws Socket, dataInBuffer chan []byte) chan error {
// 	temp := make(chan error, 1)

// 	go func() {
// 		defer close(temp)
// 		temp <- startListeningLoop(ctx, dataInBuffer, ws)
// 	}()
// 	return temp
// }

// func startListeningLoop(ctx context.Context, tunnel chan []byte, ws Socket) error {
// 	startRead := time.Now()
// 	defer func() {
// 		log.InfoF("After %d seconds, read Message over", time.Now().Unix()-startRead.Unix())
// 	}()
// 	for {
// 		log.InfoF("start read message ...")
// 		_, data, err := ws.ReadMessage()
// 		if err != nil {
// 			// endRead = time.Now()()
// 			// log.InfoF("After %d seconds, read Message over", endRead-startRead)
// 			return err
// 		}
// 		select {
// 		case tunnel <- data:
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		}
// 	}
// }

// // WritePump pumps messages from the buffer to the websocket connection.
// func (c *Connection) writePump(skt Socket, cancelSocket context.CancelFunc) (cancel context.CancelFunc) {
// 	ticker := time.NewTicker(c.pingPeriod)
// 	isCanceled := false // if server close socket because of another device login
// 	defer func() {
// 		ticker.Stop()
// 		c.invokeCallbacks(EventSocketWriteError)
// 		closeSocket(skt, isCanceled, c.uid)
// 		cancelSocket() // stop socket in http request
// 	}()
// 	var ctx context.Context
// 	ctx, cancel = context.WithCancel(context.Background())
// 	for {
// 		select {
// 		case message, ok := <-c.sendMsgBuffer:
// 			if !ok { //服务端主动关闭
// 				return
// 			}
// 			if err := writeSocket(skt, websocket.TextMessage, message, c.writeDuration); err != nil {
// 				return
// 			}
// 		case <-ticker.C:
// 			if err := writeSocket(skt, websocket.PingMessage, []byte{}, c.writeDuration); err != nil {
// 				return
// 			}
// 		case <-ctx.Done():
// 			isCanceled = true
// 			return
// 		case c.bridgeForSockets <- true: // test if online
// 		}
// 	}
// }

// func closeSocket(skt Socket, closedForLoginOnOtherDevice bool, id string) {
// 	var closeMsg []byte
// 	if closedForLoginOnOtherDevice {
// 		log.InfoF("kick off user %s ", id)
// 		closeMsg = websocket.FormatCloseMessage(websocket.CloseNormalClosure, communication.ProtoCloseLoginOnOtherDevice)
// 	} else {
// 		closeMsg = []byte{}
// 		log.InfoF("user %s leave", id)
// 	}

// 	writeSocket(skt, websocket.CloseMessage, closeMsg, 1*time.Second)
// 	skt.Close()
// }

// // AddEventCallback will add a handler for connection event
// func (c *Connection) AddEventCallback(callback func(string)) {
// 	if c.eventCallbacks == nil {
// 		c.eventCallbacks = []func(string){}
// 	}
// 	c.eventCallbacks = append(c.eventCallbacks, callback)
// }

// // write writes a message with the given message type and payload.
// func writeSocket(skt Socket, mt int, payload []byte, writeDuration time.Duration) error {
// 	if skt == nil {
// 		return nil
// 	}
// 	skt.SetWriteDeadline(time.Now().Add(writeDuration))
// 	return skt.WriteMessage(mt, payload)
// 	// return errors.New("no phisical connection")
// }

// // write writes a message with the given message type and payload.
// func (c *Connection) invokeCallbacks(evt string) {
// 	if c.eventCallbacks == nil {
// 		return
// 	}
// 	for _, callback := range c.eventCallbacks {
// 		go callback(evt)
// 	}
// }
