package main

import (
	"xsbPro/chat/node/controllers"
	log "xsbPro/log"

	"github.com/gin-gonic/gin"
	"github.com/ssor/config"

	"flag"
)

var (
	configFile    = flag.String("config", "conf/config.json", "config file for system")
	listeningPort = flag.String("port", "80", "listeningPort")

	initConf config.IConfigInfo
)

func main() {

	flag.Parse()
	if flag.Parsed() == false {
		flag.PrintDefaults()
		return
		// panic("*** para error ***")
	}

	conf, err := config.LoadConfig(*configFile)
	if err != nil {
		panic("配置文件加载错误: " + err.Error())
	}

	initConf = conf

	log.InfoF("%s", conf)

	conf.Set("nodeWanHost", conf.Get("nodeWanHost").(string)+":"+*listeningPort)
	conf.Set("nodeLanHost", conf.Get("nodeLanHost").(string)+":"+*listeningPort)

	if conf.Get("mode").(string) == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Static("/javascripts", "static/js")
	router.Static("/images", "static/img")
	router.Static("/stylesheets", "static/css")

	for _, dirMap := range controllers.ResourceDirMapList {
		router.Static(dirMap.URI, dirMap.Local)
	}
	router.LoadHTMLGlob("views/*.tpl")

	router.GET("/", controllers.LoginIndex)
	// router.GET("/", controllers.HomeDirect)
	router.GET("/connect", controllers.HomeIndirect)
	router.GET("/status", controllers.GetRunningStatus)
	router.GET("/statusIndex", controllers.StatusIndex)
	router.GET("/ws", controllers.ServeWs)
	// router.OPTIONS("/ws", controllers.ServeWsOptions)
	router.POST("/uploadImage", controllers.UploadImage)
	router.POST("/uploadAudio", controllers.UploadAudio)
	// router.GET("/groupChanged", controllers.GroupChanged)
	router.GET("/datarefresh", controllers.DataRefresh)
	// router.GET("/wslog", controllers.ServeWslog)
	// router.GET("/addGroup", controllers.AddGroup)
	router.GET("/echo", controllers.Echo)

	// go updateConfig()

	controllers.Init(conf)
	router.Run(":" + *listeningPort)
}

// func updateConfig() {
// 	// ticker := time.NewTicker(5 * time.Second)
// 	ticker := time.NewTicker(60 * time.Second)
// 	for {
// 		<-ticker.C
// 		conf, err := common.LoadConfig(*configFile)
// 		if err != nil {
// 			log.LogToFile("配置文件加载错误: " + err.Error())
// 		}
// 		if conf.GetGroupLoadCapability() != initConf.GetGroupLoadCapability() {
// 			log.InfoF("node %s(%s) capability changed to %d",
// 				conf.GetNodeLanHost(), conf.GetNodeWanHost(), conf.GetGroupLoadCapability())
// 			controllers.UpdateCapacity(conf.GetGroupLoadCapability())
// 			// initConf.NodeCapability = conf.NodeCapability
// 			initConf.Set("nodeCapability", conf.GetGroupLoadCapability())
// 		} else {
// 			log.TraceF("no change for capability %d", conf.GetGroupLoadCapability())
// 		}
// 	}
// }
