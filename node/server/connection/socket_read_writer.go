package connection

import (
	"context"
	"errors"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ssor/chat/node/server/communication"
	"github.com/ssor/log"
)

var (
	ErrSockerError = errors.New("sockerError")
)

type SocketReadWriter struct {
	socket              Socket
	pingPeriod          time.Duration
	comesInMsgBuffer    chan []byte // msg tunnel for msg from client to send to upstream
	sendMsgBuffer       chan []byte // msg tunnel for msg from upstream to sent to client
	dataBufferUnsafe    chan []byte // cache data to send to client, but not 100% successful
	onlineTest          chan bool   // used to test if socket is running
	errTunnel           chan error  // if there is an error, the error will be transported through this channel
	cancelReadWriteLoop context.CancelFunc
}

func NewSocketReadWriter(conn Socket, sendMsgBuffer chan []byte, ping time.Duration) *SocketReadWriter {
	srw := &SocketReadWriter{
		socket:           conn,
		pingPeriod:       ping,
		comesInMsgBuffer: make(chan []byte, 1),
		dataBufferUnsafe: make(chan []byte, 1024),
		sendMsgBuffer:    sendMsgBuffer,
		errTunnel:        make(chan error, 1),
		onlineTest:       make(chan bool),
	}
	srw.run()
	return srw
}

func (srw *SocketReadWriter) Online() <-chan bool {
	return srw.onlineTest
}
func (srw *SocketReadWriter) run() {
	ctx, cancel := context.WithCancel(context.Background())
	srw.cancelReadWriteLoop = cancel
	go srw.writePump(ctx)
	go srw.lanchSocketListening(ctx)
}

func (srw *SocketReadWriter) Err() <-chan error {
	return srw.errTunnel
}

func (srw *SocketReadWriter) NewData() <-chan []byte {
	return srw.comesInMsgBuffer
}

// SendDataUnsafe caches data in a channel which will be sent to client, but not 100%
func (srw *SocketReadWriter) SendDataUnsafe(data []byte) {
	srw.dataBufferUnsafe <- data
}

func (srw *SocketReadWriter) Release() {
	cancel := srw.cancelReadWriteLoop
	if cancel != nil {
		log.TraceF("release read writer...")
		cancel()
	}
}

func (srw *SocketReadWriter) lanchSocketListening(ctx context.Context) error {
	temp := make(chan error, 1)
	go func() {
		defer close(temp)
		temp <- srw.startListeningLoop(ctx)
	}()
	select {
	case err := <-temp:
		srw.errTunnel <- err
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (srw *SocketReadWriter) startListeningLoop(ctx context.Context) error {
	log.InfoF("start read message ...")
	for {
		startRead := time.Now().Unix()
		_, data, err := srw.socket.ReadMessage()
		if err != nil {
			endRead := time.Now().Unix()
			log.InfoF("After %d seconds, read Message over", endRead-startRead)
			return ErrSockerError
		}
		select {
		case srw.comesInMsgBuffer <- data:
			log.TraceF("data <- %s", string(data))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// WritePump pumps messages from the buffer to the websocket connection.
func (srw *SocketReadWriter) writePump(ctx context.Context) {
	ticker := time.NewTicker(srw.pingPeriod)
	defer func() {
		ticker.Stop()
		// spew.Dump(ctx.Err())
		srw.close(ctx.Err() == context.Canceled)
	}()
	for {
		select {
		case message, ok := <-srw.sendMsgBuffer:
			if !ok { //服务端主动关闭
				return
			}
			if !srw.sendDataCheck(message) {
				return
			}

			// if err := srw.write(websocket.TextMessage, message, 1*time.Second); err != nil {
			// 	srw.errTunnel <- ErrSockerError
			// 	return
			// }
			// log.TraceF("-> data: %s", string(message))
		case message, ok := <-srw.dataBufferUnsafe:
			if !ok { //服务端主动关闭
				return
			}
			if !srw.sendDataCheck(message) {
				return
			}
			// if err := srw.write(websocket.TextMessage, message, 1*time.Second); err != nil {
			// 	srw.errTunnel <- ErrSockerError
			// 	return
			// }
			// log.TraceF("-> data: %s", string(message))
		case <-ticker.C:
			if err := srw.write(websocket.PingMessage, []byte{}, 1*time.Second); err != nil {
				srw.errTunnel <- ErrSockerError
				return
			}
		case <-ctx.Done():
			return
		case srw.onlineTest <- true: // test if online
		}
	}
}

func (srw *SocketReadWriter) sendDataCheck(data []byte) bool {
	if err := srw.write(websocket.TextMessage, data, 1*time.Second); err != nil {
		srw.errTunnel <- ErrSockerError
		return false
	}
	log.TraceF("-> data: %s", string(data))
	return true
}
func (srw *SocketReadWriter) close(closedForLoginOnOtherDevice bool) {
	var closeMsg []byte
	if closedForLoginOnOtherDevice {
		log.InfoF("kick off user ")
		closeMsg = websocket.FormatCloseMessage(websocket.CloseNormalClosure, communication.ProtoCloseLoginOnOtherDevice)
	} else {
		closeMsg = []byte{}
		log.InfoF("user leave")
	}

	srw.write(websocket.CloseMessage, closeMsg, 1*time.Second)
	srw.socket.Close()
}

// write writes a message with the given message type and payload.
func (srw *SocketReadWriter) write(mt int, payload []byte, writeDuration time.Duration) error {
	srw.socket.SetWriteDeadline(time.Now().Add(writeDuration))
	return srw.socket.WriteMessage(mt, payload)
}
