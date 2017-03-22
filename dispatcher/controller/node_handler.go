package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"xsbPro/chat/dispatcher/dispatcher"
	"xsbPro/chat/dispatcher/resource"
	"xsbPro/chat/lua"
	"xsbPro/log"

	"github.com/gin-gonic/gin"
)

func GetNodesInfo(c *gin.Context) {
	nodes, err := dispatcher.GetNodeInfoList(nil, resource.RedisInstance.DoScript)
	if err != nil {
		log.SysF("err: %s", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nodes)
}

func GetGroupsInfo(c *gin.Context) {
	node := c.Query("node") // lan

	groups, err := lua.GetGroupsOnNode(node, resource.RedisInstance.DoScript)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, groups)
}

// CheckNodeRegistered 节点会检查自己是否登记在案
func CheckNodeRegistered(c *gin.Context) {
	node := c.Query("node")

	exists := lua.NodeExists(node, resource.RedisInstance.RedisDo)
	if exists {
		c.JSON(http.StatusOK, "OK")
	} else {
		log.TraceF("node %s registerd failed", node)
		c.JSON(http.StatusOK, "FAILED")
	}
}

//NewNodeOnLine 新节点上线时进行注册
func NewNodeOnLine(c *gin.Context) {
	type nodeInfo struct {
		Lan        string `json:"lan" binding:"required"`
		Wan        string `json:"wan" binding:"required"`
		Capability int    `json:"capability" binding:"required"`
	}

	var ni nodeInfo
	var err error

	err = c.BindJSON(&ni)
	if err != nil {
		log.SysF("RemoveArticleFromCategory error: %s", err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	// //首先检查之前是否使用同一个 IP 和端口注册过,如果有,需要删除之前的分配信息,防止不同的实例 node 使用同一个 IP 和端口登录
	// err = removeNode(ni.Lan)
	// if err != nil {
	// 	log.SysF("RemoveArticleFromCategory error: %s", err.Error())
	// 	c.JSON(http.StatusBadRequest, err.Error())
	// 	return
	// }

	err = dispatcher.RegisterToNodeCenter(dispatcher.NewNodeInfo(ni.Lan, ni.Wan, ni.Capability), resource.RedisInstance.RedisDo)
	if err != nil {
		log.SysF("RegisterToNodeCenter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = lua.UpdateNodeCapacity(ni.Lan, ni.Capability, resource.RedisInstance.DoScript)
	if err != nil {
		log.SysF("NodeUpdateCapacity error: %s", err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	log.TraceF("node %s -> %s (%d) registerd OK", ni.Lan, ni.Wan, ni.Capability)
	c.JSON(http.StatusOK, nil)
}

// NodeUpdateCapacity 更新承载量
func NodeUpdateCapacity(c *gin.Context) {
	type requestInfo struct {
		Node     string `json:"node" binding:"required"`
		Capacity int    `json:"capability" binding:"required"`
	}

	var ni requestInfo
	var err error

	err = c.BindJSON(&ni)
	if err != nil {
		log.SysF("NodeUpdateCapacity error: %s", err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = lua.UpdateNodeCapacity(ni.Node, ni.Capacity, resource.RedisInstance.DoScript)
	if err != nil {
		log.SysF("NodeUpdateCapacity error: %s", err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
}

//LoginInfoRequest ServeWs handles websocket requests from the peer.
func LoginInfoRequest(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "X-Requested-With,X_Requested_With")
	// log.TraceF("new login info request: %s", c.Request.URL)

	userID := c.Query("id")
	groupID := c.Query("group")
	if len(groupID) <= 0 || len(userID) <= 0 {
		log.InfoF("no user ID: [%s] group ID: [%s]", userID, groupID)
		c.JSON(http.StatusOK, errors.New("no user ID or group ID"))
		return
	}

	wan, err := lua.GetNodeByGroup(groupID, resource.RedisInstance.DoScript)
	if err != nil {
		log.SysF("LoginInfoRequest error: %s", err)
		c.JSON(http.StatusOK, newResponse(1, "system error", nil))
		return
	}
	//已分配到指定节点
	if len(wan) > 0 {
		url := fmt.Sprintf("/ws?group=%s&id=%s", groupID, userID)
		hosts := []string{"ws://" + wan}
		// c.JSON(http.StatusOK, NewResponse(0, "", NewLoginInfo(url, hosts)))
		bs, err := jsonMarshal(newResponse(0, "", NewLoginInfo(url, hosts)), true)
		if err != nil {
			log.SysF("LoginInfoRequest error: %s", err)
			c.JSON(http.StatusOK, newResponse(404, "", NewLoginInfo("", []string{})))
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", bs)
		return
	}
	//尚未找到分配节点
	log.TraceF("no node for group %s yet", groupID)
	c.JSON(http.StatusOK, newResponse(404, "", NewLoginInfo("", []string{})))
	return
}

func jsonMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

// func join_group_to_node_map(group string) string {
// 	// return string_group_node_map + group
// 	// return fmt.Sprintf(common.String_group_node_map, group, "*")
// 	return format_group_node_map_key(group, "*", "*")
// }

// //通过 group 和 node 的对应,找到 group 分配到的节点
// func get_node_by_group_map(group string) (lan, wan string, e error) {

// 	keys, err := redis.Strings(Redis_instance.RedisDo("keys", join_group_to_node_map(group)))
// 	if err != nil {
// 		if err != redis.ErrNil {
// 			log.SysF("error: ", err)
// 			e = err
// 			return
// 		} else {
// 			return
// 		}
// 	}

// 	//将 host 从 字符串中解析出来 "group->%s,node->%s,wan->%s"
// 	if len(keys) > 0 {
// 		log.TraceF("keys: %s", keys)
// 		group_node_wan := strings.Split(keys[0], ",")
// 		if len(group_node_wan) != 3 {
// 			return "", "", errors.New("key format no right")
// 		}
// 		node_host := strings.Split(group_node_wan[1], "->")
// 		if len(node_host) != 2 {
// 			return "", "", errors.New("key format no right")
// 		}
// 		lan = node_host[1]
// 		wan_host := strings.Split(group_node_wan[2], "->")
// 		if len(wan_host) != 2 {
// 			return "", "", errors.New("key format no right")
// 		}
// 		wan = wan_host[1]

// 		return
// 	}
// 	return
// }

// //NodeDispatchRequest 节点主要要求承载支部
// func NodeDispatchRequest(c *gin.Context) {
// 	type requestInfo struct {
// 		Node  string `json:"node" binding:"required"`
// 		Group string `json:"group" binding:"required"`
// 	}

// 	var ni requestInfo
// 	var err error

// 	err = c.BindJSON(&ni)
// 	if err != nil {
// 		log.SysF("NodeDispatchRequest error: %s", err.Error())
// 		c.JSON(http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	node_manager.newGroupRequest(ni.Group, ni.Node)
// 	c.JSON(http.StatusOK, nil)
// }
