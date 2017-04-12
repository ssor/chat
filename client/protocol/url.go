package protocol

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/parnurzeal/gorequest"
)

// FormatConnectURL returns url format for client connecting to chat server
func FormatConnectURL(host, id, group string) string {
	// url := fmt.Sprintf("ws://%s/ws?id=%s&group=%s", host, id, group)
	url := fmt.Sprintf("%s/ws?id=%s&group=%s", host, id, group)
	return url
}

// FormatLoginURL returns url for logging to dispather
func FormatLoginURL(host, id, group string) string {
	url := fmt.Sprintf("http://%s/login?id=%s&group=%s", host, id, group)
	return url
}

// RequestNodeHost returns host of node for group
func RequestNodeHost(dispatcherHost, group string) (host string, err error) {
	_, res, errs := gorequest.New().Get(FormatLoginURL(dispatcherHost, "test", group)).End()
	if errs != nil {
		err = errs[0]
		return
	}
	var resMessage struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
		Data struct {
			Hosts []string `json:"hosts"`
		} `json:"data"`
	}
	err = json.Unmarshal([]byte(res), &resMessage)
	if err != nil {
		return
	}
	if resMessage.Code != 0 {
		err = errors.New(resMessage.Msg)
		return
	}
	if resMessage.Data.Hosts == nil || len(resMessage.Data.Hosts) <= 0 {
		err = errors.New("no host")
		return
	}

	host = resMessage.Data.Hosts[0]
	return
}
