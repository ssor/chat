package lua

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ssor/chat/lua/scripts"
	"github.com/ssor/chat/mongo"
	"github.com/ssor/chat/redis"
	"github.com/ssor/log"
	redigo "github.com/ssor/redigo/redis"
)

// ScriptExecutor : type for execute script
type ScriptExecutor func(script *redis.Script, keysAndArgs ...interface{}) (interface{}, error)

// RedisDo :
type RedisDo func(cmd string, args ...interface{}) (interface{}, error)

//GetAllGroups will
func GetAllGroups(scriptExecutor ScriptExecutor) (interface{}, error) {
	args := redigo.Args{}
	return scriptExecutor(luaScripts.Scripts[luaScriptGetAllGroups], args...)
}

//FillNewGroupToRedis 要将 group 置为 非删除状态
func FillNewGroupToRedis(group *mongo.Group, scriptExecutor ScriptExecutor) error {
	args := redigo.Args{}.Add(scripts.ForamtGroupKey(group.ID)).AddFlat(group)
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptFillNewGroup], args...)
	if err != nil {
		log.SysF("FillNewGroupToRedis error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("FillNewGroupToRedis failed")
	}
	return nil
}

//ResetGroupsStatusInRedis 将 redis 中的 group 置为删除状态
func ResetGroupsStatusInRedis(scriptExecutor ScriptExecutor) error {
	args := redigo.Args{}
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptResetGroupsStatus], args...)
	if err != nil {
		log.SysF("ResetGroupsStatusInRedis error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("ResetGroupsStatusInRedis failed")
	}
	return nil
}

//AdjustNodeDispatch 由于 group 信息发生变化,对 node 的承载进行微调,尤其原来分配到 node 上的 group 已经被删除的情况下
func AdjustNodeDispatch(scriptExecutor ScriptExecutor) error {
	args := redigo.Args{}
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptAdjustNodeDispatch], args...)
	if err != nil {
		log.SysF("AdjustNodeDispatch error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("AdjustNodeDispatch failed")
	}
	return nil
}

//FillGroupsToRedis  will
func FillGroupsToRedis(groups []*mongo.Group, scriptExecutor ScriptExecutor) error {

	for _, group := range groups {
		err := FillNewGroupToRedis(group, scriptExecutor)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveUsersFromRedis will
func RemoveUsersFromRedis(users []string, cmdsExecutor func(*redis.Commands) error) error {

	cmds := redis.NewCommands(true)

	for _, user := range users {
		cmds.Add("DEL", redigo.Args{}.Add(scripts.FormatUserKey(user)))
	}

	err := cmdsExecutor(cmds)
	return err
}

//ClearUsersInRedis clear users info in redis
func ClearUsersInRedis(scriptExecutor ScriptExecutor) error {
	args := redigo.Args{}
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptClearUsers], args...)
	// res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_clear_users])
	if err != nil {
		log.SysF("ClearUsersInRedis error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("ClearUsersInRedis failed")
	}
	return nil
}

// FillUsersToRedis will
func FillUsersToRedis(users mongo.UserArray, cmdsExecutor func(*redis.Commands) error) error {

	cmds := redis.NewCommands(true)

	for _, user := range users {
		// cmds.Add("HMSET", redis.Args{}.Add(lua.Get_user_key(user.ID)).AddFlat(user)...)
		cmds.Add("HMSET", redigo.Args{}.Add(scripts.FormatUserKey(user.ID)).AddFlat(user)...)
	}

	err := cmdsExecutor(cmds)
	return err
}

// FillGroupUserRelationshipToRedis will
func FillGroupUserRelationshipToRedis(group string, users []string, cmdsExecutor func(*redis.Commands) error) error {

	cmds := redis.NewCommands(true)

	key := scripts.FormatGroupUserRelationKey(group)
	cmds.Add("DEL", redigo.Args{}.Add(key))

	args := redigo.Args{}.Add(key)
	for _, user := range users {
		args = args.AddFlat(user)
	}

	cmds.Add("SADD", args...)
	err := cmdsExecutor(cmds)
	return err
}

// func RemoveGroupFromRedis(group string, scriptExecutor ScriptExecutor) error {

// RemoveGroup remove group in redis cache
func RemoveGroup(group string, scriptExecutor ScriptExecutor) error {
	// args := redis.Args{}.Add(fmt.Sprintf(key_format_group, group))
	args := redigo.Args{}.Add(scripts.ForamtGroupKey)
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptRemoveGroup], args...)
	if err != nil {
		log.SysF("RemoveGroupFromRedis error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("RemoveGroupFromRedis failed")
	}
	return nil
}

// func GetGroupUsersFromCache(group string, scriptExecutor ScriptExecutor) (db.UserArray, error) {

// GetGroupUsers return users in group
func GetGroupUsers(group string, scriptExecutor ScriptExecutor) (mongo.UserArray, error) {
	args := redigo.Args{}.AddFlat(group)
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptGetUsersInGroup], args...)
	if err != nil {
		log.InfoF("GetGroupUsersFromCache DoScript err: %s", err)
		return nil, err
	}
	data := res.([]uint8)

	if string(data) == "{}" {
		return mongo.UserArray{}, nil
	}

	return unmarshalUsers(data), nil
}

func unmarshalUsers(data []byte) mongo.UserArray {

	type UserCache struct {
		ID            string `bson:"_id" redis:"id"`
		Password      string `bson:"password" redis:"password"`
		Name          string `bson:"name" redis:"name"`
		Phone         string `bson:"phone" redis:"phone"`
		Email         string `bson:"email" redis:"email"`
		Group         string `bson:"group" redis:"group"`
		Index         string `bson:"index" redis:"index"`
		Actived       string `bson:"actived" redis:"actived"`
		Image         string `bson:"image" redis:"image"`
		Tag           string `bson:"tag" redis:"tag"`
		Chief         string `bson:"chief" redis:"chief"`
		Gender        string `bson:"gender" redis:"gender"`
		Department    string `bson:"department" redis:"department"`
		UserType      string `bson:"userType" redis:"userType"`
		ActiveCode    string `bson:"activeCode" redis:"activeCode"`
		LastLoginTime string `bson:"lastLoginTime" redis:"-"`
		LastLoginAddr string `bson:"lastLoginAddr" redis:"-"`
		JoinPartyDate string `bson:"joinPartyDate" redis:"joinPartyDate"`
		BirthDay      string `bson:"birthDay" redis:"birthDay"`
		Title         string `bson:"title" redis:"title"`
		Community     string `bson:"community" redis:"community"`
	}
	var usersCache []*UserCache
	err := json.Unmarshal(data, &usersCache)
	if err != nil {
		log.Info(string(data))
		log.InfoF("unmarshalUsers  err: %s", err)
		return nil
	}

	convertUserCacheToUser := func(cache *UserCache) *mongo.User {
		index, err := strconv.Atoi(cache.Index)
		if err != nil {
			index = 0
		}
		actived, err := strconv.Atoi(cache.Actived)
		if err != nil {
			actived = 0
		}
		gender, err := strconv.Atoi(cache.Gender)
		if err != nil {
			gender = 0
		}
		userType, err := strconv.Atoi(cache.UserType)
		if err != nil {
			userType = 0
		}
		chief, err := strconv.ParseBool(cache.Chief)
		if err != nil {
			chief = false
		}
		user := mongo.NewUser(cache.ID, cache.Phone, cache.Email, cache.Password, cache.Name, "",
			cache.Group, cache.Image, cache.Department, cache.Tag, gender, index, chief, userType, cache.ActiveCode, time.Now().Format(time.RFC3339))
		user.Actived = actived
		return user
	}
	users := mongo.UserArray{}
	for _, cache := range usersCache {
		users = append(users, convertUserCacheToUser(cache))
	}
	return users
}

//GetNodeByGroup 通过 group 和 node 的对应,找到 group 分配到的节点
func GetNodeByGroup(group string, scriptExecutor ScriptExecutor) (wan string, e error) {
	args := redigo.Args{}.Add(scripts.ForamtGroupKey(group))
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptGetNodeByGroup], args...)
	if err != nil {
		return "", err
	}

	return string(res.([]uint8)), nil
}

// GetGroupsOnNode return all groups on node
func GetGroupsOnNode(lan string, scriptExecutor ScriptExecutor) ([]*mongo.Group, error) {
	// node_key := Format_nodeinfo_key(lan)
	nodeKey := scripts.FormatNodeInfoKey(lan)
	args := redigo.Args{}.Add(nodeKey)
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptGetGroupsOnNode], args...)
	if err != nil {
		log.SysF("GetGroupsOnNode error: %s", err)
		return nil, err
	}
	var groups []*mongo.Group
	err = json.Unmarshal(res.([]uint8), &groups)
	if err != nil {
		log.InfoF("data: %s", string(res.([]uint8)))
		return nil, err
	}
	return groups, nil
}

// GetAllNodes will
func GetAllNodes(scriptExecutor ScriptExecutor) (interface{}, error) {
	return scriptExecutor(luaScripts.Scripts[luaScriptGetAllNodes], redigo.Args{})
}

// GetGroupUsersOnNode return users of group
// the group must be dispatched to specified node, or no data should be returned
func GetGroupUsersOnNode(group, nodeLan string, scriptExecutor ScriptExecutor) (mongo.UserArray, error) {

	args := redigo.Args{}.AddFlat(scripts.ForamtGroupKey(group)).AddFlat(scripts.FormatNodeInfoKey(nodeLan)).AddFlat(group)
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptGetGroupUsersOnNode], args...)
	if err != nil {
		log.SysF("GetGroupUsersOnNode error: %s", err)
		return nil, err
	}

	data := res.([]uint8)
	if string(data) == "[]" {
		return nil, nil
	}
	return unmarshalUsers(data), nil
	// var users db.UserArray
	// err = json.Unmarshal(res.([]uint8), &users)
	// if err != nil {
	// 	log.SysF("GetGroupUsersOnNode res: %s", string(res.([]uint8)))
	// 	return nil, err
	// }
	// return users, nil

}

//GetNodeInfo will
func GetNodeInfo(nodeID string, scriptExecutor ScriptExecutor) (interface{}, error) {
	nodeKey := scripts.FormatNodeInfoKey(nodeID)
	return scriptExecutor(luaScripts.Scripts[luaScriptGetNodeInfo], redigo.Args{}.Add(nodeKey)...)
}

// GetAllNodeKeys will
func GetAllNodeKeys(redisDo RedisDo) ([]string, error) {
	nodeKeys, err := redigo.Strings(redisDo("keys", scripts.FormatNodeInfoKey("*")))
	if err != nil {
		if err == redigo.ErrNil {
			return []string{}, nil
		}
		return nil, err
	}

	return nodeKeys, nil
}

// NodeExists will
func NodeExists(lan string, redisDo RedisDo) bool {

	res, err := redisDo("EXISTS", redigo.Args{}.Add(scripts.FormatNodeInfoKey(lan))...)
	if err != nil {
		log.SysF("NodeExists error: %s", err)
		return false
	}
	if res.(int64) == 1 {
		return true
	}
	return false
}

// RemoveNode will
func RemoveNode(lan string, scriptExecutor ScriptExecutor) error {
	nodeKey := scripts.FormatNodeInfoKey(lan)
	args := redigo.Args{}.Add(nodeKey)
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptRemoveNode], args...)
	if err != nil {
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("RemoveNode failed")
	}
	return nil
}

// UpdateNodeCapacity will
func UpdateNodeCapacity(ip string, cap int, scriptExecutor ScriptExecutor) error {
	if len(ip) > 0 {
		nodeKey := scripts.FormatNodeInfoKey(ip)
		args := redigo.Args{}.Add(nodeKey).AddFlat(cap)
		res, err := scriptExecutor(luaScripts.Scripts[luaScriptUpdateNodeCapability], args...)
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

//GetUnloadGroupCount will
func GetUnloadGroupCount(scriptExecutor ScriptExecutor) (int, error) {
	res, err := scriptExecutor(luaScripts.Scripts[luaScriptGetUnloadGroupCount], redigo.Args{})
	if err != nil {
		return 0, err
	}

	return int(res.(int64)), nil
}

func GetNodeInfoByKey(nodeKey string, scriptExecutor ScriptExecutor) ([]byte, error) {
	var err error
	res, err := scriptExecutor(luaScripts.Scripts[scripts.GetNodeInfo], redigo.Args{}.Add(nodeKey)...)
	if err != nil {
		log.SysF("GetNodeInfoByKey error: %s", err)
		return nil, err
	}
	if res == nil {
		return nil, errors.New("noData")
	}
	return res.([]byte), nil
}

// type luaFunc struct {
// 	Content    string
// 	OutsideRef []interface{}
// }

// func (lf *luaFunc) fill() string {
// 	if lf.OutsideRef != nil &&
// 		len(lf.OutsideRef) > 0 {
// 		return fmt.Sprintf(lf.Content, lf.OutsideRef...)
// 	}
// 	return lf.Content
// }

// func newLuaFunc(content string, outsideRefs ...interface{}) *luaFunc {
// 	return &luaFunc{
// 		Content:    content,
// 		OutsideRef: outsideRefs,
// 	}
// }

// var (
// 	luaFuncGetUsersInGroup      = newLuaFunc(scripts.GetUsersInGroup, key_format_group_user_relation, key_format_user)
// 	LuaFuncRes2Table            = newLuaFunc(scripts.Res2table)
// 	LuaFuncAdjustNodeDispatch   = newLuaFunc(scripts.AdjustNodeDispatch)
// 	LuaFuncClearUsers           = newLuaFunc(scripts.ClearUsers)
// 	LuaFuncFillNewGroup         = newLuaFunc(scripts.FillNewGroup)
// 	LuaFuncFindNodeNotFull      = newLuaFunc(scripts.FindNodeNotFull)
// 	LuaFuncGetAllGroups         = newLuaFunc(scripts.GetAllGroups)
// 	LuaFuncGetAllNodes          = newLuaFunc(scripts.GetAllNodes)
// 	LuaFuncGetGroupUsersOnNode  = newLuaFunc(scripts.GetGroupUsersOnNode)
// 	LuaFuncGetGroupsOnNode      = newLuaFunc(scripts.GetGroupsOnNode)
// 	LuaFuncGetGroupsUnlinked    = newLuaFunc(scripts.GetGroupsUnlinked)
// 	LuaFuncGetNodeByGroup       = newLuaFunc(scripts.GetNodeByGroup)
// 	LuaFuncGetNodeInfo          = newLuaFunc(scripts.GetNodeInfo)
// 	LuaFuncGetUnloadGroupCount  = newLuaFunc(scripts.GetUnloadGroupCount)
// 	LuaFuncGetUsersInGroup      = newLuaFunc(scripts.GetUsersInGroup)
// 	LuaFuncPrintLog             = newLuaFunc(scripts.PrintLog)
// 	LuaFuncRemoveGroup          = newLuaFunc(scripts.RemoveGroup)
// 	LuaFuncRemoveNode           = newLuaFunc(scripts.RemoveNode)
// 	LuaFuncResetGroupStatus     = newLuaFunc(scripts.ResetGroupsStatus)
// 	LuaFuncSetNodeLinkToGroup   = newLuaFunc(scripts.SetNodeLinkToGroup)
// 	LuaFuncString2Table         = newLuaFunc(scripts.StringToTable)
// 	LuaFuncTableLength          = newLuaFunc(scripts.TableLength)
// 	LuaFuncUpdateNodeCapability = newLuaFunc(scripts.UpdateNodeCapability)
// )

var (

// lua_func_search_group_name = fmt.Sprintf(`
// 	local function search_group_name(group_name)
// 		print_log("to search " .. group_name .. " --->>>")
// 		local groups = {}
// 		local keys_of_group = redis.call("keys", "group:*")
// 		for index, key_of_group in ipairs(keys_of_group) do
// 			local res_group = redis.call("hgetall", key_of_group)
// 			if res_group then
// 				local group = res2table(res_group)
// 				local find_result = string.find(group.name, group_name)
// 				if find_result then
// 					table.insert(groups, group)
// 				else
// 					print_log("group " .. group.id .. " name is " .. group.name .. " outside results")
// 				end
// 			end
// 		end
// 		return groups
// 	end
// `)
)
