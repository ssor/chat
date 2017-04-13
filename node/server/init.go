package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/ssor/config"
	"github.com/ssor/log"
)

var (
	// debugMode = false
	debugMode = true
)

var (
	maxUserUnactiveDuration   = 24 * time.Hour     //user 最长非活动时长,超过该时长则需要确认该 user 是否还在支部中
	maxChatLogReserveDuration = 24 * 3 * time.Hour //聊天记录最长保存时间
	// interval_check_node_registered_status       = 1 * time.Minute
	defaultMaxMemory int64 = 5 << 20 // 32 MB

	uploadStaticImageFileURL string
	uploadStaticAudioFileURL string

	// hubManager     *HubManager
	serverInstance *server

	nodeID = ""
)

func init() {
	serverInstance = newServer()
	// hubManager = NewHubManager()
	// go hubManager.Run()
}

// Init set paras for node to run
func Init(conf config.IConfigInfo, node string) {
	uploadStaticImageFileURL = conf.Get("staticServerHost").(string) + "/api/v1/image/upload?type=chat&uid=%s&para=%s"
	uploadStaticAudioFileURL = conf.Get("staticServerHost").(string) + "/api/v1/audio/upload?uid=%s&para=%s"

	nodeID = node
}

// //向 node 管理中心注册
func registerToNodeCenter(center, lan, wan string, loadCapability int) error {
	type nodeInfo struct {
		Lan        string `json:"lan" binding:"required"`
		Wan        string `json:"wan" binding:"required"`
		Capability int    `json:"capability" binding:"required"`
	}

	url := fmt.Sprintf("http://%s/nodeOnLine", center)
	res, _, errs := gorequest.New().Post(url).Send(nodeInfo{Lan: lan, Wan: wan, Capability: loadCapability}).End()
	if errs != nil && len(errs) > 0 {
		log.InfoF("registerToNodeCenter error:%s", errs[0])
		return errs[0]
	}
	if res.StatusCode != http.StatusOK {
		log.SysF("dispatchGroup status: %s", res.Status)
		return errors.New("node register failed")
	}
	return nil
}
