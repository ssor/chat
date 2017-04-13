package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/ssor/chat/node/resource"
	"github.com/ssor/chat/node/server"
	"github.com/ssor/config"
	"github.com/ssor/log"
)

var (
	// debugMode = false
	debugMode = true
)

var (
	conf               config.IConfigInfo
	ResourceDirMapList = make(map[string]*ResourceDirMap)

	isCenterAlive = false // 如果 dispatcher 主动监测 node 存在,说明 dispatcher 是活动的,那么 node 也同时可以确定 dispatcher 是活动的,通过这个思路可以减少互相监测造成的资源浪费
)

var (
	shareImagePath = "/shareImage/"
	shareAudioPath = "/shareAudio/"
	imageDir       = "chatFiles/images"
	audioDir       = "chatFiles/audioes"

	// upload_static_image_file_url string
	// upload_static_audio_file_url string

	// max_user_unactive_duration            = 24 * time.Hour     //user 最长非活动时长,超过该时长则需要确认该 user 是否还在支部中
	// max_chat_log_reserve_duration         = 24 * 3 * time.Hour //聊天记录最长保存时间
	interval_check_node_registered_status = 1 * time.Minute

	// dispatch_request_group_cache chan string
	// hubManager                   *models.HubManager
	// redis_instance *common.Redis_Instance
	// mongo_pool *MongoSessionPool
)

func init() {
}

func Init(_conf config.IConfigInfo) {
	conf = _conf
	ResourceDirMapList[imageDir] = newResourceDirMap(imageDir, shareImagePath, conf.Get("domain").(string))
	ResourceDirMapList[audioDir] = newResourceDirMap(audioDir, shareAudioPath, conf.Get("domain").(string))

	initAppDir()

	// upload_static_image_file_url = _conf.GetStaticFileServer() + "/api/v1/image/upload?type=chat&uid=%s&para=%s"
	// upload_static_audio_file_url = _conf.GetStaticFileServer() + "/api/v1/audio/upload?uid=%s&para=%s"

	resource.Init(_conf)
	if resource.RedisInstance == nil {
		panic("redis is nil after init")
	}

	server.Init(_conf, _conf.Get("nodeLanHost").(string))
	err := registerToNodeCenter(conf.Get("registerCenterHost").(string), conf.Get("nodeLanHost").(string), conf.Get("nodeWanHost").(string), conf.Get("nodeCapability").(int))
	if err != nil {
		log.SysF("registerToNodeCenter err: %s", err)
	}
	go keepNodeRegisteredInCenter()
}
func keepNodeRegisteredInCenter() {
	// ticker := time.NewTicker(10 * time.Second) //debug
	ticker := time.NewTicker(interval_check_node_registered_status)
	for {
		<-ticker.C
		if isCenterAlive == true {
			isCenterAlive = false
			continue
		}
		registerCenterHost := conf.Get("registerCenterHost").(string)
		nodeLanHost := conf.Get("nodeLanHost").(string)
		nodeWanHost := conf.Get("nodeWanHost").(string)
		nodeCapability := conf.Get("nodeCapability").(int)
		if checkNodeRegisteredInCenter(registerCenterHost, nodeLanHost) == false {
			err := registerToNodeCenter(registerCenterHost, nodeLanHost, nodeWanHost, nodeCapability)
			if err != nil {
				log.SysF("registerToNodeCenter err: %s", err)
			}
		}
	}
}

// UpdateCapacity update cap of node
func UpdateCapacity(cap int) {
	type requestInfo struct {
		Node     string `json:"node" binding:"required"`
		Capacity int    `json:"capability" binding:"required"`
	}
	go func() {
		log.InfoF("node %s capacity updated to %d", conf.Get("nodeLanHost").(string), cap)
		url := fmt.Sprintf("http://%s/nodeUpdateCapacity", conf.Get("registerCenterHost"))
		_, _, errs := gorequest.New().Post(url).Send(requestInfo{Node: conf.Get("nodeLanHost").(string), Capacity: cap}).End()
		if errs != nil && len(errs) > 0 {
			log.InfoF("requestGroupDispatchLoop error:%s", errs[0])
		}
	}()
}

func checkNodeRegisteredInCenter(center, host string) bool {
	url := fmt.Sprintf("http://%s/checkNodeRegistered?node=%s", center, host)
	res, body, errs := gorequest.New().Get(url).End()
	if errs != nil && len(errs) > 0 {
		log.SysF("checkNodeRegisteredInCenter error: %s", errs[0])
		return false
	}
	// log.TraceF("body : %s", body)
	if res.StatusCode == http.StatusOK {
		if strings.Contains(body, "FAILED") {
			log.SysF("node %s not registered", host)
			return false
			// panic("node should restart")
		}
	} else {
		return false
	}
	return true
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

//创建程序运行的必需基础目录
func initAppDir() {
	for dir := range ResourceDirMapList {

		if b, _ := fileExists(dir); b == true {
			continue
		}

		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	// //cp 一下系统需要的文件
	// filepath.Walk(data_dir, func(path string, info os.FileInfo, err error) error {

	// 	if strings.HasPrefix(info.Name(), ".") {
	// 		return nil
	// 	}

	// 	if info.IsDir() {
	// 		fmt.Println("DIR : ", path, " -> ", info.Name())
	// 	} else {
	// 		fmt.Println("FILE: ", path, " -> ", info.Name())

	// 		newPath := strings.Replace(path, data_dir, Base_dir, 1)
	// 		if b, _ := exists(newPath); b == true {
	// 			return nil
	// 		}
	// 		_, err := filecopy(path, newPath)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// })
}

// fileExists returns whether the given file or directory exists or not
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func getGoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func stack() []byte {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return buf[:n]
}
