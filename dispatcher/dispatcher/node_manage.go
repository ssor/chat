package dispatcher

import (
	"errors"
	"strings"
	"xsbPro/chatDispatcher/lua"
	"xsbPro/log"

	"fmt"

	"github.com/parnurzeal/gorequest"
	"github.com/ssor/redigo/redis"
)

const (
	node_state_alive = "alive"
	node_state_dead  = "dead"
)

//
//NodeManager 负责处理 group 分配请求,将 group 分配到已注册的节点
//同时定期检查节点状态,并清除不活跃节点
//
type RedisDo func(cmd string, args ...interface{}) (interface{}, error)

var (
	err_no_node                   = errors.New("no node to dispatch")
	err_all_node_full_load        = errors.New("all node full load")
	err_record_did_too_many_times = errors.New("too many times did on this record")
)

func NodeExists(lan string, redisDo func(cmd string, args ...interface{}) (interface{}, error)) bool {

	res, err := redisDo("EXISTS", redis.Args{}.Add(lua.Format_nodeinfo_key(lan))...)
	if err != nil {
		log.SysF("NodeExists error: %s", err)
		return false
	}
	if res.(int64) == 1 {
		return true
	}
	return false
}

//向 node 管理中心注册,如果注册失败则退出
func RegisterToNodeCenter(node_info *NodeInfo, redisDo func(cmd string, args ...interface{}) (interface{}, error)) error {
	_, err := redisDo("hmset", redis.Args{}.Add(node_info.Key).AddFlat(node_info)...)
	if err != nil {
		log.SysF("registerToNodeCenter error: %s", err)
		return err
	}
	return nil
}

func RemoveNode(lan string, scriptExecutor ScriptExecutor) error {
	node_key := lua.Format_nodeinfo_key(lan)
	args := redis.Args{}.Add(node_key)
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_remove_node], args...)
	if err != nil {
		log.SysF("RemoveNode error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("RemoveNode failed")
	}
	return nil
}

// func (nm *NodeManager) updateNodeCapacity(ip string, cap int) {
func UpdateNodeCapacity(ip string, cap int, scriptExecutor ScriptExecutor) error {
	if len(ip) > 0 {

		node_key := lua.Format_nodeinfo_key(ip)
		args := redis.Args{}.Add(node_key).AddFlat(cap)
		res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_update_node_capability], args...)
		if err != nil {
			log.SysF("updateNodeCapacity error: %s", err)
			return err
		}
		if string(res.([]uint8)) != "OK" {
			return fmt.Errorf("update cap failed")
		}
		log.InfoF("node %s capacity updated to %d", ip, cap)
	}
	return nil
}

func GetUnloadGroupCount(scriptExecutor ScriptExecutor) (int, error) {
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_unload_group_count], redis.Args{})
	if err != nil {
		log.SysF("GetUnloadGroupCount error: %s", err)
		return 0, err
	}

	return int(res.(int64)), nil
}

// //如果指定节点,则分配到该节点
// func (nm *NodeManager) newGroupRequest(group, lan string) {
// 	if len(lan) > 0 {
// 		nm.new_group_request <- &NodeOpRecord{Op: op_dispatch_group_to_node, Paras: []interface{}{group, lan}}
// 	} else {
// 		nm.new_group_request <- &NodeOpRecord{Op: op_dispatch_group, Paras: []interface{}{group}}
// 	}
// }

//if node registered is not alive any more, clear it
func CheckNodeState(redisDo RedisDo, next func(string, string) error) {
	log.Info("CheckNodeState --->>>")
	node_keys, err := getAllNodeKeys(redisDo)
	if err != nil {
		log.InfoF("CheckNodeState error: %s", err)
		return
	}
	log.InfoF("get %d nodes now", len(node_keys))
	for _, key := range node_keys {
		split_result := strings.Split(key, "->")
		if len(split_result) < 2 {
			log.SysF("sys error: node format error")
			return
		}
		lan := split_result[1]
		state := checkNodeState(lan)
		log.InfoF("state of node %s now is %s", lan, state)
		if next != nil {
			err := next(state, lan)
			if err != nil {
				log.SysF("sys error: %s", err.Error())
			}
		}
		// if state == node_state_dead {
		// 	err := next(lan)
		// 	log.SysF("sys error: %s", err)
		// }
	}
}

func checkNodeState(ip string) string {
	res, _, errs := gorequest.New().Get(getEchoUrl(ip)).End()
	if errs != nil && len(errs) > 0 {
		log.InfoF("checkNodeState error:%s", errs[0])
		return node_state_dead
	}
	if res.StatusCode != 200 {
		log.InfoF("node %s die", ip)
		return node_state_dead
	}
	// log.TraceF("node %s running well", ip)

	return node_state_alive
}

// func (nm *NodeManager) run() {
// 	// go nm.startNodeOpRecordHandler()

// 	tickerCheckNodeState := time.NewTicker(node_alive_check_interval)
// 	for {
// 		select {
// 		// case record := <-nm.new_group_request:
// 		// 	log.TraceF("new group %s comes in", record.Paras[0])
// 		// 	nm.add_op <- record
// 		case <-tickerCheckNodeState.C:
// 			nm.checkNodeState()
// 		}
// 	}
// }

// //按顺序单线程执行各个操作
// func (nm *NodeManager) startNodeOpRecordHandler() {
// 	ticker_normal := time.NewTicker(500 * time.Millisecond)
// 	ticker_exception := time.NewTicker(5 * time.Second)
// 	for {
// 		select {
// 		case op := <-nm.add_op:
// 			nm.op_record_new = append(nm.op_record_new, op)
// 		case <-ticker_normal.C:
// 			// if no err, remove it; or move it to op_record_exception
// 			if len(nm.op_record_new) > 0 {
// 				first_record := nm.op_record_new[0]
// 				nm.op_record_new = nm.op_record_new[1:]
// 				err := doRecord(first_record)
// 				if err != nil {
// 					log.InfoF("do record %s failed: %s", first_record, err)
// 					nm.op_record_exception = append(nm.op_record_exception, first_record)
// 				}
// 			}
// 		case <-ticker_exception.C:
// 			// if no err or if too many times, remove it; else wait for next time
// 			if len(nm.op_record_exception) > 0 {
// 				first_record := nm.op_record_exception[0]
// 				err := doRecord(first_record)
// 				if err == nil {
// 					nm.op_record_exception = nm.op_record_exception[1:]
// 				} else if err == err_record_did_too_many_times {
// 					log.InfoF("record %s exeed max count", first_record)
// 					nm.op_record_exception = nm.op_record_exception[1:]
// 				} else {
// 					log.InfoF("do record %s failed: %s", first_record, err)
// 				}
// 			}
// 		}
// 	}
// }

// func doRecord(record *NodeOpRecord) error {
// 	if record.ExecuteCount > max_count_did_on_record {
// 		return err_record_did_too_many_times
// 	}
// 	var err error
// 	switch record.Op {
// 	case op_dispatch_group:
// 		err = dispatchGroup(record.Paras[0].(string))
// 	case op_dispatch_group_to_node:
// 		err = dispatchGroupToNode(record.Paras)
// 	case op_remove_node:
// 		err = removeNode(record.Paras[0].(string))
// 	case op_update_node_capacity:
// 		err = updateNodeCapacity(record.Paras)
// 	}
// 	//记录执行次数
// 	record.ExecuteCount++
// 	return err
// }

// func UpdateNodeCapacity(paras []interface{}) error {
// 	ip := paras[0].(string)
// 	cap := paras[1].(int)
// 	ni, err := getNodeInfoByKey(format_nodeinfo_key(ip))
// 	if err != nil {
// 		log.InfoF("dispatchGroupToNode error: %s", err)
// 		return err
// 	}
// 	if ni == nil {
// 		log.InfoF("node %s unregistered but want to update capacity")
// 		return nil
// 	}

// 	ni.Max = cap
// 	cmds := common.NewRedisCommands(true)
// 	cmds.Add("hmset", redis.Args{}.Add(ni.Key).AddFlat(ni)...)
// 	err = Redis_instance.RedisDoMulti(cmds)
// 	if err != nil {
// 		log.SysF("updateNodeCapacity error: %s", err)
// 		return err
// 	}
// 	log.InfoF("node %s capacity updated to %d", ip, cap)
// 	return nil
// }

// //paras: {group,ip}
// func dispatchGroupToNode(paras []interface{}) error {
// 	if len(paras) < 2 {
// 		return nil
// 	}
// 	group := paras[0].(string)
// 	lan := paras[1].(string)
// 	// wan := paras[2].(string)

// 	//首先需要确认是否已经分配
// 	lan, _, err := get_node_by_group_map(group)
// 	if err != nil {
// 		log.InfoF("dispatchGroupToNode error: %s", err)
// 		return err
// 	}
// 	if len(lan) > 0 {
// 		//说明已经分配过了,无论是否指定都不再分配
// 		return nil
// 	}

// 	ni, err := getNodeInfoByKey(format_nodeinfo_key(lan))
// 	if err != nil {
// 		log.InfoF("dispatchGroupToNode error: %s", err)
// 		return err
// 	}
// 	if ni == nil {
// 		return err_all_node_full_load
// 	}
// 	return write_dispatch_to_redis(ni, group)
// 	// //通知 node 准备支部数据
// 	// res, _, errs := gorequest.New().Get(getAddGroupUrl(lan, group)).End()
// 	// if errs != nil && len(errs) > 0 {
// 	// 	log.InfoF("dispatchGroupToNode error:%s", errs[0])
// 	// 	return errs[0]
// 	// }
// 	// if res.StatusCode != http.StatusOK {
// 	// 	log.SysF("dispatchGroupToNode status: %s", res.Status)
// 	// 	return errors.New("node prepare group failed")
// 	// }

// 	// //将结果写入到 redis 中
// 	// cmds := common.NewRedisCommands(true)
// 	// cmds.Add("set", format_group_node_map_key(group, lan, wan), wan)
// 	// ni.Current += 1
// 	// cmds.Add("hmset", redis.Args{}.Add(ni.Key).AddFlat(ni)...)
// 	// // cmds.Add("sadd", redis.Args{}.Add(set_node_groups_map+ni.IP).AddFlat(group)...)
// 	// err = redis_instance.RedisDoMulti(cmds)
// 	// if err != nil {
// 	// 	log.SysF("dispatchGroupToNode error: %s", err)
// 	// 	return err
// 	// }

// 	// log.TraceF("dispatch group %s to node %s -> %s", group, ni.LanHost, ni.WanHost)
// 	// return nil
// }

// func dispatchGroup(group string) error {
// 	//首先需要确认是否已经分配
// 	lan, _, err := get_node_by_group_map(group)
// 	if err != nil {
// 		log.InfoF("dispatchGroup error: %s", err)
// 		return err
// 	}
// 	if len(lan) > 0 {
// 		return nil
// 	}

// 	ni, err := getSingleNodeInfo(func(node *NodeInfo) bool {
// 		return node.Max > node.Current
// 	})
// 	// ni, err := getNodeNotFullLoad(node_keys)
// 	if err != nil {
// 		log.InfoF("dispatchGroup error: %s", err)
// 		return err
// 	}
// 	if ni == nil {
// 		return err_all_node_full_load
// 	}

// 	return write_dispatch_to_redis(ni, group)

// 	// //通知 node 准备支部数据

// 	// res, _, errs := gorequest.New().Get(getAddGroupUrl(ni.LanHost, group)).End()
// 	// if errs != nil && len(errs) > 0 {
// 	// 	log.InfoF("dispatchGroup error:%s", errs[0])
// 	// 	return errs[0]
// 	// }
// 	// if res.StatusCode != http.StatusOK {
// 	// 	log.SysF("dispatchGroup status: %s", res.Status)
// 	// 	return errors.New("node prepare group failed")
// 	// }

// 	// //将结果写入到 redis 中
// 	// cmds := common.NewRedisCommands(true)
// 	// cmds.Add("set", format_group_node_map_key(group, ni.LanHost, ni.WanHost), ni.WanHost)
// 	// // cmds.Add("set", fmt.Sprintf(common.String_group_node_map, group, ni.IP), ni.IP)
// 	// // cmds.Add("set", string_group_node_map+group, ni.IP)
// 	// ni.Current += 1
// 	// cmds.Add("hmset", redis.Args{}.Add(ni.Key).AddFlat(ni)...)
// 	// // cmds.Add("sadd", redis.Args{}.Add(set_node_groups_map+ni.IP).AddFlat(group)...)
// 	// err = redis_instance.RedisDoMulti(cmds)
// 	// if err != nil {
// 	// 	log.SysF("dispatchGroup error: %s", err)
// 	// 	return err
// 	// }

// 	// log.TraceF("dispatch group %s to node %s -> %s", group, ni.LanHost, ni.WanHost)
// 	// return nil
// }

// func write_dispatch_to_redis(ni *NodeInfo, group string) error {

// 	//通知 node 准备支部数据

// 	res, _, errs := gorequest.New().Get(getAddGroupUrl(ni.LanHost, group)).End()
// 	if errs != nil && len(errs) > 0 {
// 		log.InfoF("dispatchGroup error:%s", errs[0])
// 		return errs[0]
// 	}
// 	if res.StatusCode != http.StatusOK {
// 		log.SysF("dispatchGroup status: %s", res.Status)
// 		return errors.New("node prepare group failed")
// 	}

// 	//将结果写入到 redis 中
// 	cmds := common.NewRedisCommands(true)
// 	cmds.Add("set", format_group_node_map_key(group, ni.LanHost, ni.WanHost), ni.WanHost)
// 	// cmds.Add("set", fmt.Sprintf(common.String_group_node_map, group, ni.IP), ni.IP)
// 	// cmds.Add("set", string_group_node_map+group, ni.IP)
// 	ni.Current += 1
// 	cmds.Add("hmset", redis.Args{}.Add(ni.Key).AddFlat(ni)...)
// 	// cmds.Add("sadd", redis.Args{}.Add(set_node_groups_map+ni.IP).AddFlat(group)...)
// 	err := Redis_instance.RedisDoMulti(cmds)
// 	if err != nil {
// 		log.SysF("dispatchGroup error: %s", err)
// 		return err
// 	}

// 	log.TraceF("dispatch group %s to node %s -> %s", group, ni.LanHost, ni.WanHost)
// 	return nil
// }
