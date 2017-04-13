package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ssor/chat/node/resource"
	"github.com/ssor/chat/node/server"
	"github.com/ssor/log"
)

var (
	// defaultMaxMemory   int64 = 5 << 20 // 32 MB
	chatFilesPathImage = "chatFiles/images/"
	chatFilesPathAudio = "chatFiles/audioes/"
)

var (
	Datarefresh_update_users_of_group = "update_users_of_group"
	Datarefresh_remove_group          = "remove_group"
	Datarefresh_refresh_all_data      = "refreshalldata"
)

func DataRefresh(c *gin.Context) {

	opt := c.Query("opt")
	para := c.Query("para")
	var err error
	switch opt {
	//所有数据全部更新
	case Datarefresh_refresh_all_data:
		err = server.RefreshAll(resource.RedisInstance.DoScript)
		//移除具体支部
	case Datarefresh_remove_group:
		err = server.RemoveGroup(para)
	//更新支部内成员信息
	case Datarefresh_update_users_of_group:
		err = server.RefreshGroupUsers(para, resource.RedisInstance.DoScript)
	default:
	}

	if err != nil {
		log.SysF("err: %s", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}

// func GroupChanged(c *gin.Context) {
// 	changedType := c.Query("type")
// 	groupID := c.Query("group")

// 	err := server.RefreshGroup(changedType, groupID, resource.Redis_instance.DoScript)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, nil)
// }

func GetRunningStatus(c *gin.Context) {
	sr := server.NodeStatusReport()
	c.JSON(http.StatusOK, sr)
}

func Echo(c *gin.Context) {
	c.Header("Connection", "close")

	isCenterAlive = true

	c.JSON(http.StatusOK, nil)
}

// func AddGroup(c *gin.Context) {

// 	groupID := c.Query("group")
// 	if len(groupID) <= 0 {
// 		log.InfoF("no group ID: [%s]", groupID)
// 		c.JSON(http.StatusBadRequest, errors.New("no group ID"))
// 		return
// 	}
// 	hm := hubManager
// 	if hm == nil {
// 		log.InfoF("no group ID: [%s]", groupID)
// 		c.JSON(http.StatusBadRequest, errors.New("no group ID"))
// 		return
// 	}
// 	hub := hm.Hubs.Get(groupID)
// 	if hub == nil {
// 		//创建新聊天组
// 		users := getUsersInGroup(groupID)
// 		if users == nil {
// 			log.SysF("系统异常, 无法获取支部中的用户信息")
// 			c.JSON(http.StatusBadRequest, errors.New("支部错误,或者该支部中尚未添加用户"))
// 			return
// 		}

// 		hub = models.NewHub(groupID, users)
// 		hm.Add(hub)
// 		log.TraceF("new hub %s has %d users", groupID, users.Length())
// 	}

// 	c.JSON(http.StatusOK, nil)
// }
