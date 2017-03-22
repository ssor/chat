package controller

// "fmt"

// "strings"
// "net/url"

func newResponse(code int, message string, data interface{}) *response {
	return &response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
