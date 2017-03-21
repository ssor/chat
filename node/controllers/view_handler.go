package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//Home use
func HomeDirect(c *gin.Context) {
	c.HTML(http.StatusOK, "direct.tpl", gin.H{"HOST": c.Request.Host})
}

func HomeIndirect(c *gin.Context) {
	userID := c.Query("user")
	groupID := c.Query("group")
	c.HTML(http.StatusOK, "indirect.tpl", gin.H{"HOST": conf.GetNodeWanHost(), "USER": userID, "GROUP": groupID})
	// c.HTML(http.StatusOK, "indirect.tpl", gin.H{"HOST": conf.GetRegisterCenterHost(), "USER": userID, "GROUP": groupID})
}

func LoginIndex(c *gin.Context) {
	group := c.Query("group")
	c.HTML(http.StatusOK, "login.tpl", gin.H{"GROUP": group})
}

func StatusIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "status.tpl", nil)
}
