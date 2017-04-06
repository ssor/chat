package protocol

import "fmt"

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
