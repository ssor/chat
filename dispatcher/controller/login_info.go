package controller

// "fmt"

//LoginInfo 返回给客户端的登录信息
type LoginInfo struct {
	URL   string   `json:"url"`
	Hosts []string `json:"hosts"`
	// Host string
}

// NewLoginInfo init a login info
func NewLoginInfo(url string, hosts []string) *LoginInfo {
	return &LoginInfo{
		URL:   url,
		Hosts: hosts,
	}
}
