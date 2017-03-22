package dispatcher

import (
	"fmt"
	"xsbPro/log"

	"encoding/json"

	"strconv"

	"xsbPro/chat/lua"
	"xsbPro/chat/lua/scripts"

	"github.com/davecgh/go-spew/spew"
)

//node 需要实现以下接口
const (
	nodeHandlerEcho        = "http://%s/echo"
	nodeHandlerDataRefresh = "http://%s/datarefresh?opt=%s&para=%s"
	// node_handler_refreshall   = "http://%s/alldatarefresh"
	// node_handler_add_group    = "http://%s/addGroup?group=%s"
)

var (
	DataRefreshUpdateUsersOfGroup = "update_users_of_group"
	DataRefreshRemoveGroup        = "remove_group"
	DataRefreshRefreshAllData     = "refreshalldata"
)

type NodeInfo struct {
	Key     string `redis:"-" json:"-"`
	LanHost string `redis:"lan" json:"lan"` //作为区分节点的标识,同时用于检测节点的存活
	WanHost string `redis:"wan" json:"wan"`
	Max     int    `redis:"-" json:"capacity"`
	Current int    `redis:"current" json:"current"`
}

func (ni *NodeInfo) Equal(nodeInfo *NodeInfo) bool {
	if nodeInfo == nil {
		return false
	}

	if ni.LanHost != nodeInfo.LanHost ||
		ni.WanHost != nodeInfo.WanHost {
		return false
	}
	return true
}

func NewNodeInfo(lan, wan string, max int) *NodeInfo {
	return &NodeInfo{
		Key:     scripts.FormatNodeInfoKey(lan),
		LanHost: lan,
		WanHost: wan,
		Max:     max,
		Current: 0,
	}
}

func formatNodeDataRefreshDataURL(ip, opt, para string) string {
	return fmt.Sprintf(nodeHandlerDataRefresh, ip, opt, para)
}

// func getAddGroupUrl(ip, group string) string {
// 	return fmt.Sprintf(node_handler_add_group, ip, group)
// }

func formatEchoURL(ip string) string {
	return fmt.Sprintf(nodeHandlerEcho, ip)
}

// func getRefreshAllDataUrl(ip string) string {
// 	return fmt.Sprintf(node_handler_refreshall, ip)
// }

func GetNodeInfoList(p func(*NodeInfo) bool, scriptExecutor ScriptExecutor) ([]*NodeInfo, error) {

	// res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_all_nodes], redis.Args{})
	res, err := lua.GetAllNodes(lua.ScriptExecutor(scriptExecutor))
	if err != nil {
		log.SysF("GetNodeInfoList error: %s", err)
		return nil, err
	}

	var nodesRaw []*struct {
		Key     string `json:"-"`
		LanHost string `json:"lan"` //作为区分节点的标识,同时用于检测节点的存活
		WanHost string `json:"wan"`
		Max     string `json:"max"`
		Current string `json:"current"`
	}
	bs := res.([]uint8)
	if len(bs) <= 2 {
		return nil, nil
	}
	err = json.Unmarshal(bs, &nodesRaw)
	if err != nil {
		spew.Dump(string(bs))
		return nil, err
	}
	nodes := []*NodeInfo{}
	for _, nodeRaw := range nodesRaw {
		max := 0
		current := 0
		if len(nodeRaw.Max) > 0 {
			max, err = strconv.Atoi(nodeRaw.Max)
			if err != nil {
				return nil, fmt.Errorf("max err: %s", err)
			}
		}
		if len(nodeRaw.Current) > 0 {
			current, err = strconv.Atoi(nodeRaw.Current)
			if err != nil {
				return nil, fmt.Errorf("current err: %s", err)
			}
		}
		node := NewNodeInfo(nodeRaw.LanHost, nodeRaw.WanHost, max)
		node.Current = current
		if p != nil {
			if p(node) == true {
				nodes = append(nodes, node)
			}
		} else {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// //通过 group 和 node 的对应,找到 group 分配到的节点
// func GetNodeByGroup(group string, scriptExecutor ScriptExecutor) (wan string, e error) {
// 	args := redis.Args{}.Add(lua.Get_group_key(group))
// 	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_node_by_group], args...)
// 	if err != nil {
// 		log.SysF("GetNodeByGroup error: %s", err)
// 		return "", err
// 	}

// 	return string(res.([]uint8)), nil
// }

// func GetNodeInfoByKey(node_key string, scriptExecutor ScriptExecutor) (*NodeInfo, error) {

// 	var err error
// 	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_nodeinfo], redis.Args{}.Add(node_key)...)
// 	if err != nil {
// 		log.SysF("GetNodeInfoByKey error: %s", err)
// 		return nil, err
// 	}

// 	var ni struct {
// 		LanHost string `json:"lan"` //作为区分节点的标识,同时用于检测节点的存活
// 		WanHost string `json:"wan"`
// 		Max     string `json:"capacity"`
// 		Current string `json:"current"`
// 	}
// 	err = json.Unmarshal(res.([]uint8), &ni)
// 	if err != nil {
// 		log.InfoF("GetNodeInfoByKey => %s ", string(res.([]uint8)))
// 		return nil, err
// 	}
// 	var max, current int
// 	if len(ni.Max) > 0 {
// 		max, err = strconv.Atoi(ni.Max)
// 		if err != nil {
// 			log.InfoF("GetNodeInfoByKey => %s ", string(res.([]uint8)))
// 			return nil, err
// 		}
// 	}
// 	current, err = strconv.Atoi(ni.Current)
// 	if err != nil {
// 		log.InfoF("GetNodeInfoByKey => %s ", string(res.([]uint8)))
// 		return nil, err
// 	}
// 	nni := NewNodeInfo(ni.LanHost, ni.WanHost, max)
// 	nni.Current = current
// 	return nni, nil
// }

// func getAllNodeKeys(redisDo RedisDo) ([]string, error) {
// 	node_keys, err := redis.Strings(redisDo("keys", lua.Format_nodeinfo_key("*")))
// 	if err != nil {
// 		if err == redis.ErrNil {
// 			return []string{}, nil
// 		}
// 		log.SysF("getAllNodeKeys error: %s", err)
// 		return nil, err
// 	}

// 	return node_keys, nil
// }
