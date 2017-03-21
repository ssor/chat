package dispatcher

import (
	"fmt"
	"xsbPro/chatDispatcher/lua"
	"xsbPro/log"

	"encoding/json"

	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/ssor/redigo/redis"
)

//node 需要实现以下接口
const (
	node_handler_echo         = "http://%s/echo"
	node_handler_data_refresh = "http://%s/datarefresh?opt=%s&para=%s"
	// node_handler_refreshall   = "http://%s/alldatarefresh"
	// node_handler_add_group    = "http://%s/addGroup?group=%s"
)

var (
	Datarefresh_update_users_of_group = "update_users_of_group"
	Datarefresh_remove_group          = "remove_group"
	Datarefresh_refresh_all_data      = "refreshalldata"
)

type NodeInfo struct {
	Key     string `redis:"-" json:"-"`
	LanHost string `redis:"lan" json:"lan"` //作为区分节点的标识,同时用于检测节点的存活
	WanHost string `redis:"wan" json:"wan"`
	Max     int    `redis:"-" json:"capacity"`
	Current int    `redis:"current" json:"current"`
}

func (ni *NodeInfo) Equal(node_info *NodeInfo) bool {
	if node_info == nil {
		return false
	}

	if ni.LanHost != node_info.LanHost ||
		ni.WanHost != node_info.WanHost {
		return false
	}
	return true
}

func NewNodeInfo(lan, wan string, max int) *NodeInfo {
	return &NodeInfo{
		Key:     lua.Format_nodeinfo_key(lan),
		LanHost: lan,
		WanHost: wan,
		Max:     max,
		Current: 0,
	}
}

func getNodeDataRefreshDataUrl(ip, opt, para string) string {
	return fmt.Sprintf(node_handler_data_refresh, ip, opt, para)
}

// func getAddGroupUrl(ip, group string) string {
// 	return fmt.Sprintf(node_handler_add_group, ip, group)
// }

func getEchoUrl(ip string) string {
	return fmt.Sprintf(node_handler_echo, ip)
}

// func getRefreshAllDataUrl(ip string) string {
// 	return fmt.Sprintf(node_handler_refreshall, ip)
// }

func GetNodeInfoList(p func(*NodeInfo) bool, scriptExecutor ScriptExecutor) ([]*NodeInfo, error) {

	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_all_nodes], redis.Args{})
	if err != nil {
		log.SysF("GetNodeInfoList error: %s", err)
		return nil, err
	}

	var nodes_raw []*struct {
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
	err = json.Unmarshal(bs, &nodes_raw)
	if err != nil {
		spew.Dump(string(bs))
		return nil, err
	}
	nodes := []*NodeInfo{}
	for _, node_raw := range nodes_raw {
		max := 0
		current := 0
		if len(node_raw.Max) > 0 {
			max, err = strconv.Atoi(node_raw.Max)
			if err != nil {
				return nil, fmt.Errorf("max err: %s", err)
			}
		}
		if len(node_raw.Current) > 0 {
			current, err = strconv.Atoi(node_raw.Current)
			if err != nil {
				return nil, fmt.Errorf("current err: %s", err)
			}
		}
		node := NewNodeInfo(node_raw.LanHost, node_raw.WanHost, max)
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

//通过 group 和 node 的对应,找到 group 分配到的节点
func GetNodeByGroup(group string, scriptExecutor ScriptExecutor) (wan string, e error) {
	args := redis.Args{}.Add(lua.Get_group_key(group))
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_node_by_group], args...)
	if err != nil {
		log.SysF("GetNodeByGroup error: %s", err)
		return "", err
	}

	return string(res.([]uint8)), nil
}

func GetNodeInfoByKey(node_key string, scriptExecutor ScriptExecutor) (*NodeInfo, error) {

	var err error
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_nodeinfo], redis.Args{}.Add(node_key)...)
	if err != nil {
		log.SysF("GetNodeInfoByKey error: %s", err)
		return nil, err
	}

	var ni struct {
		LanHost string `json:"lan"` //作为区分节点的标识,同时用于检测节点的存活
		WanHost string `json:"wan"`
		Max     string `json:"capacity"`
		Current string `json:"current"`
	}
	err = json.Unmarshal(res.([]uint8), &ni)
	if err != nil {
		log.InfoF("GetNodeInfoByKey => %s ", string(res.([]uint8)))
		return nil, err
	}
	var max, current int
	if len(ni.Max) > 0 {
		max, err = strconv.Atoi(ni.Max)
		if err != nil {
			log.InfoF("GetNodeInfoByKey => %s ", string(res.([]uint8)))
			return nil, err
		}
	}
	current, err = strconv.Atoi(ni.Current)
	if err != nil {
		log.InfoF("GetNodeInfoByKey => %s ", string(res.([]uint8)))
		return nil, err
	}
	nni := NewNodeInfo(ni.LanHost, ni.WanHost, max)
	nni.Current = current
	return nni, nil
}

func getAllNodeKeys(redisDo RedisDo) ([]string, error) {
	node_keys, err := redis.Strings(redisDo("keys", lua.Format_nodeinfo_key("*")))
	if err != nil {
		if err == redis.ErrNil {
			return []string{}, nil
		}
		log.SysF("getAllNodeKeys error: %s", err)
		return nil, err
	}

	return node_keys, nil
}
