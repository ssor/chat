package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

//返回给客户端的登录信息
type LoginInfo struct {
	Url   string   `json:"url"`
	Hosts []string `json:"hosts"`
	// Host string
}

func NewLoginInfo(url string, hosts []string) *LoginInfo {
	return &LoginInfo{
		Url:   url,
		Hosts: hosts,
	}
}

func JSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}
func TestUrlEncoding(t *testing.T) {
	Convey("TestUrlEncoding ", t, func() {

		url := fmt.Sprintf(`/ws?group=%s&id=%s`, "123", "abc")
		hosts := []string{"ws://"}
		res := NewResponse(0, "", NewLoginInfo(url, hosts))
		bs, err := JSONMarshal(res, true)
		Panic_error(err)
		So(len(bs), ShouldNotEqual, 0)
		So(strings.Contains(string(bs), `\u0026`), ShouldEqual, false)
	})
}

func TestSplitKey(t *testing.T) {
	Convey("TestSplitKey", t, func() {

		key := `group->1459828066_1459951271326749291,node->172.16.1.35:8082,wan->dev.chat.dyxsb.net:8082`
		group_node_wan := strings.Split(key, ",")
		So(len(group_node_wan), ShouldEqual, 3)
		node_host := strings.Split(group_node_wan[1], "->")
		So(len(node_host), ShouldEqual, 2)
		wan_host := strings.Split(group_node_wan[2], "->")
		So(len(wan_host), ShouldEqual, 2)
	})
}

func Panic_error(err error) {
	if err != nil {
		panic(err.Error())
	}
}
