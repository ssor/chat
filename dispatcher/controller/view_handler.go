package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ssor/chat/dispatcher/dispatcher"
	"github.com/ssor/chat/dispatcher/resource"
	"github.com/ssor/log"
)

func Nodes(c *gin.Context) {
	c.HTML(http.StatusOK, "nodes.tpl", gin.H{"HOST": c.Request.Host})
}

func Groups(c *gin.Context) {
	node := c.Query("node")
	c.HTML(http.StatusOK, "groups.tpl", gin.H{"NODE": node})
}

func SearchIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "search.tpl", nil)
}

func Search(c *gin.Context) {
	group := c.Query("group")
	user := c.Query("user")

	if len(group) > 0 {
		log.InfoF("search => %s", group)
		results, err := dispatcher.SearchGroupName(group, resource.RedisInstance.DoScript)
		if err != nil {
			log.SysF("err: %s", err)
			c.JSON(http.StatusOK, nil)
			return
		}
		log.TraceF("got %d search results", len(results))
		c.JSON(http.StatusOK, results)
		return
	} else if len(user) > 0 {
		return
	}
	c.JSON(http.StatusOK, nil)
}
