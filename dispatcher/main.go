package main

import (
	"flag"

	"github.com/gin-gonic/gin"
	"github.com/ssor/chat/dispatcher/controller"
	"github.com/ssor/config"
)

var (
	configFile = flag.String("config", "conf/config.json", "config file for system")
)

func main() {

	flag.Parse()
	if flag.Parsed() == false {
		flag.PrintDefaults()
		return
	}

	conf, err := config.LoadConfig(*configFile)
	if err != nil {
		panic("配置文件加载错误: " + err.Error())
	}

	// if conf.Mode == "release" {
	if conf.Get("mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	controller.Init(conf)

	router := gin.Default()

	router.Static("/javascripts", "static/js")
	router.Static("/images", "static/img")
	router.Static("/stylesheets", "static/css")
	router.Static("/datatable", "static/dataTable")
	router.LoadHTMLGlob("views/*.tpl")

	// router.GET("", controller.Image)
	router.POST("/nodeOnLine", controller.NewNodeOnLine)
	router.GET("/login", controller.LoginInfoRequest)
	router.OPTIONS("/login", handleOptions)
	router.POST("/nodeUpdateCapacity", controller.NodeUpdateCapacity)
	router.GET("/checkNodeRegistered", controller.CheckNodeRegistered)
	router.GET("/groupsInfo", controller.GetGroupsInfo)
	router.GET("/nodesInfo", controller.GetNodesInfo)
	router.GET("/nodesIndex", controller.Nodes)
	router.GET("/searchIndex", controller.SearchIndex)
	router.GET("/groupsIndex", controller.Groups)
	router.GET("/search", controller.Search)
	// router.POST("/nodeDispatchRequest", controller.NodeDispatchRequest)

	// router.GET("/", controllers.Home)
	// router.GET("/ws", controllers.ServeWs)
	// router.POST("/uploadImage", controllers.UploadImage)
	// router.POST("/uploadAudio", controllers.UploadAudio)
	// router.GET("/wslog", controllers.ServeWslog)

	router.Run(":" + conf.Get("listeningPort").(string))
}

func handleOptions(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "X-Requested-With,X_Requested_With")
}
