package tests

import (
	"fmt"
	"testing"
	"time"
	"xsbPro/chatServer/controllers/models"

	. "github.com/smartystreets/goconvey/convey"
)

type ReportLog struct{}

func (rl *ReportLog) NewMessage([]byte) {
}
func (rl *ReportLog) ConnError(string) {
}

type FakeSocket struct{}

func (fs *FakeSocket) ReadMessage() (int, []byte, error) {
	return 0, []byte{}, nil
}
func (fs *FakeSocket) WriteMessage(messageType int, data []byte) error {
	return nil
}

func (fs *FakeSocket) SetWriteDeadline(t time.Time) error {
	return nil
}

func (fs *FakeSocket) Close() error {
	return nil
}

func TestModel(t *testing.T) {

	Convey("--> 最底层的 socket 连接的读写", t, func() {
		//模仿大量连接,并且每个连接同时进行写以及关闭操作
		operation := func(index int) {
			fs := &FakeSocket{}
			rl := &ReportLog{}
			conn := models.NewConnection(fs, rl, fmt.Sprintf("conn_%d", index))
			msg, _ := models.NewMessage(2, "id", "name", "")
			ch := make(chan int)
			go func() {
				<-ch
				for loop := 0; loop < 100; loop++ {
					conn.Close("", 1*time.Second)
				}
			}()
			go func() {
				<-ch
				for loop := 0; loop < 100; loop++ {
					conn.Send(models.NewMessageRecord(msg))
				}
			}()
			close(ch)
		}
		for index := 0; index < 10; index++ {
			operation(index)
		}
	})

	Convey("--> User 对消息的管理", t, func() {

	})

	Convey("--> Hub 对 User 的管理", t, func() {

	})

	Convey("--> HubManager 对 Hub 的管理", t, func() {

	})
}
