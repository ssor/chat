package dispatcher

import (
	"fmt"
	"time"
	"xsbPro/chat/dispatcher/resource"
	"xsbPro/chat/lua"
	"xsbPro/common"
	"xsbPro/log"
	db "xsbPro/xsbdb"

	"github.com/parnurzeal/gorequest"
	"github.com/ssor/config"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	conf config.IConfigInfo

	nodeAliveCheckInterval = 30 * time.Second
)

func init() {
}

// Init will fill data to redis and remind nodes to sync data
func Init(_conf config.IConfigInfo, scriptExecutor ScriptExecutor, redisDo RedisDo) {
	conf = _conf

	prepareDataInRedis(_conf)

	go func() {
		tickerCheckNodeState := time.NewTicker(nodeAliveCheckInterval)
		for {
			select {
			case <-tickerCheckNodeState.C:
				f := func(state, lan string) error {
					return removeDeadNode(state, lan, scriptExecutor)
				}
				CheckNodeState(redisDo, f)
			}
		}
	}()

	//dispather 重新填充数据后, 提醒节点进行数据同步
	go NotifyNodeDataRefresh(DataRefreshRefreshAllData, "", redisDo)
}

func removeDeadNode(state, lan string, scriptExecutor ScriptExecutor) error {
	if state == nodeStateDead {
		err := lua.RemoveNode(lan, lua.ScriptExecutor(scriptExecutor))
		if err != nil {
			return err
		}
		log.InfoF("node %s removed", lan)
	}
	return nil
}

// NotifyNodeDataRefresh will request alive node to sync data with dispatcher
func NotifyNodeDataRefresh(opt, para string, redisDo RedisDo) error {
	//remind all node to refresh data
	f := func(state, lan string) error {
		if state == nodeStateAlive {
			log.InfoF("remind node %s to refresh data", lan)
			_, _, errs := gorequest.New().Get(formatNodeDataRefreshDataURL(lan, opt, para)).End()
			if errs != nil && len(errs) > 0 {
				log.InfoF("remind node %s to refresh data error:%s", lan, errs[0])
				return fmt.Errorf("remind node to refresh data error:%s", errs[0])
			}
		}
		return nil
	}
	CheckNodeState(redisDo, f)
	return nil
}

// func ClearHistoryData(cleaner func() error) {
// 	err := cleaner()
// 	if err != nil {
// 		panic(err)
// 	}
// }

//整理
func prepareDataInRedis(_conf config.IConfigInfo) {
	// log.Info("prepareDataInRedis --->>>")
	session, err := resource.MongoPool.GetSession()
	if err != nil {
		panic("mongo conn err: " + err.Error())
	}
	defer resource.MongoPool.ReturnSession(session, err)
	err = FillDataToRedisFromMongo(session, _conf.Get("dbName").(string), resource.RedisInstance.RedisDoMulti, resource.RedisInstance.DoScript)
	if err != nil {
		panic(err)
	}
}

// //要将 group 置为 非删除状态
// func FillNewGroupToRedis(group *db.Group, scriptExecutor ScriptExecutor) error {
// 	args := redis.Args{}.Add(lua.Get_group_key(group.ID)).AddFlat(group)
// 	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_fill_new_group], args...)
// 	if err != nil {
// 		log.SysF("FillNewGroupToRedis error: %s", err)
// 		return err
// 	}
// 	if string(res.([]uint8)) != "OK" {
// 		return fmt.Errorf("FillNewGroupToRedis failed")
// 	}
// 	return nil
// }

// //将 redis 中的 group 置为删除状态
// func ResetGroupsStatusInRedis(scriptExecutor ScriptExecutor) error {
// 	args := redis.Args{}
// 	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_reset_groups_status], args...)
// 	if err != nil {
// 		log.SysF("ResetGroupsStatusInRedis error: %s", err)
// 		return err
// 	}
// 	if string(res.([]uint8)) != "OK" {
// 		return fmt.Errorf("ResetGroupsStatusInRedis failed")
// 	}
// 	return nil
// }

// //由于 group 信息发生变化,对 node 的承载进行微调,尤其原来分配到 node 上的 group 已经被删除的情况下
// func AdjustNodeDispatch(scriptExecutor ScriptExecutor) error {
// 	args := redis.Args{}
// 	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_adjust_node_dispatch], args...)
// 	if err != nil {
// 		log.SysF("AdjustNodeDispatch error: %s", err)
// 		return err
// 	}
// 	if string(res.([]uint8)) != "OK" {
// 		return fmt.Errorf("AdjustNodeDispatch failed")
// 	}
// 	return nil
// }

// func FillGroupsToRedis(groups []*db.Group, cmdsExecutor func(*common.RedisCommands) error) error {

// func FillGroupsToRedis(groups []*db.Group, scriptExecutor ScriptExecutor) error {

// 	for _, group := range groups {
// 		err := FillNewGroupToRedis(group, scriptExecutor)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func RemoveUsersFromRedis(users []string, cmdsExecutor func(*common.RedisCommands) error) error {

// 	cmds := common.NewRedisCommands(true)

// 	for _, user := range users {
// 		cmds.Add("DEL", redis.Args{}.Add(lua.Get_user_key(user)))
// 	}

// 	err := cmdsExecutor(cmds)
// 	return err
// }

// //clear users info in redis
// func ClearUsersInRedis(scriptExecutor ScriptExecutor) error {
// 	args := redis.Args{}
// 	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_clear_users], args...)
// 	// res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_clear_users])
// 	if err != nil {
// 		log.SysF("ClearUsersInRedis error: %s", err)
// 		return err
// 	}
// 	if string(res.([]uint8)) != "OK" {
// 		return fmt.Errorf("ClearUsersInRedis failed")
// 	}
// 	return nil
// }

// func FillUsersToRedis(users db.UserArray, cmdsExecutor func(*common.RedisCommands) error) error {

// 	cmds := common.NewRedisCommands(true)

// 	for _, user := range users {
// 		cmds.Add("HMSET", redis.Args{}.Add(lua.Get_user_key(user.ID)).AddFlat(user)...)
// 	}

// 	err := cmdsExecutor(cmds)
// 	return err
// }

// UpdateUsersOfGroup will
func UpdateUsersOfGroup(session *mgo.Session, dbName string, group string, cmdsExecutor func(*common.RedisCommands) error) error {
	//支部中用户信息数据重新填充
	users, err := getUsersFromDB(session, dbName, bson.M{"group": group})
	if err != nil {
		return err
	}
	getIDofUsers := func(users db.UserArray) []string {
		l := []string{}
		for _, user := range users {
			l = append(l, user.ID)
		}
		return l
	}
	err = lua.FillGroupUserRelationshipToRedis(group, getIDofUsers(users), cmdsExecutor)
	if err != nil {
		return err
	}
	return nil
}

// func FillGroupUserRelationshipToRedis(group string, users []string, cmdsExecutor func(*common.RedisCommands) error) error {

// 	cmds := common.NewRedisCommands(true)

// 	key := lua.Get_group_user_relation_key(group)
// 	cmds.Add("DEL", redis.Args{}.Add(key))

// 	args := redis.Args{}.Add(key)
// 	for _, user := range users {
// 		args = args.AddFlat(user)
// 	}
// 	cmds.Add("SADD", args...)
// 	err := cmdsExecutor(cmds)
// 	return err
// }

//FillDataToRedisFromMongo 每次载入数据,将 redis 中数据与 mongo 数据进行比对,完成后,检查分配信息,调整 node的实际承载量,需要的话,进行新的分配
func FillDataToRedisFromMongo(session *mgo.Session, dbName string, cmdsExecutor func(*common.RedisCommands) error, scriptExecutor ScriptExecutor) error {
	groups, err := getGroupsFromDB(session, dbName, nil)
	if err != nil {
		return err
	}

	err = lua.ResetGroupsStatusInRedis(lua.ScriptExecutor(scriptExecutor))
	if err != nil {
		return err
	}

	err = lua.FillGroupsToRedis(groups, lua.ScriptExecutor(scriptExecutor))
	if err != nil {
		return err
	}

	err = lua.AdjustNodeDispatch(lua.ScriptExecutor(scriptExecutor))
	if err != nil {
		return err
	}

	err = lua.ClearUsersInRedis(lua.ScriptExecutor(scriptExecutor))
	if err != nil {
		return err
	}
	//支部中用户信息数据重新填充
	users, err := getUsersFromDB(session, dbName, nil)
	if err != nil {
		return err
	}
	err = lua.FillUsersToRedis(users, cmdsExecutor)
	if err != nil {
		return err
	}

	getGroupUsers := func(group *db.Group, users db.UserArray) db.UserArray {
		usersInGroup := db.UserArray{}
		for _, user := range users {
			if user.Group == group.ID {
				usersInGroup = append(usersInGroup, user)
			}
		}
		return usersInGroup
	}
	getIDofUsers := func(users db.UserArray) []string {
		l := []string{}
		for _, user := range users {
			l = append(l, user.ID)
		}
		return l
	}
	for _, group := range groups {
		usersInGroup := getGroupUsers(group, users)
		// log.InfoF("group %s has %d users", group.ID, len(usersInGroup))
		err = lua.FillGroupUserRelationshipToRedis(group.ID, getIDofUsers(usersInGroup), cmdsExecutor)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddUsers get user info from mongo and fill it to redis
func AddUsers(session *mgo.Session, dbName string, query interface{}, cmdsExecutor func(*common.RedisCommands) error) error {
	users, err := getUsersFromDB(session, dbName, query)
	if err != nil {
		return err
	}
	err = lua.FillUsersToRedis(users, cmdsExecutor)
	if err != nil {
		return err
	}
	return nil
}

func getUsersFromDB(session *mgo.Session, dbName string, query interface{}) (db.UserArray, error) {

	if session == nil {
		return nil, fmt.Errorf("db session should not be nil")
	}

	var users db.UserArray
	err := session.DB(dbName).C(common.Collection_userinfo).Find(query).All(&users)
	if err != nil {
		log.SysF("getUsersFromDB error: %s", err)
		return nil, err
	}

	return users, nil
}

func getGroupsFromDB(session *mgo.Session, dbName string, query interface{}) ([]*db.Group, error) {

	if session == nil {
		return nil, fmt.Errorf("db session should not be nil")
	}

	var groups []*db.Group
	err := session.DB(dbName).C(common.Collection_group).Find(query).All(&groups)
	if err != nil {
		log.SysF("GetGroupsFromDB error: %s", err)
		return nil, err
	}

	return groups, nil
}

// AddNewGroup get group info including users in group and fill it to redis
func AddNewGroup(session *mgo.Session, dbName, groupID string, cmdsExecutor func(*common.RedisCommands) error, scriptExecutor ScriptExecutor) error {
	groups, err := getGroupsFromDB(session, dbName, bson.M{"_id": groupID})
	if err != nil {
		log.SysF("get group from db err: %s", err)
		return err
	}
	if len(groups) > 0 {
		err = lua.FillGroupsToRedis(groups, lua.ScriptExecutor(scriptExecutor))
		if err != nil {
			log.SysF("FillGroupsToRedis  err: %s", err)
			return err
		}
	}
	//支部中用户信息数据重新填充
	users, err := getUsersFromDB(session, dbName, bson.M{"group": groupID})
	if err != nil {
		return err
	}
	err = lua.FillUsersToRedis(users, cmdsExecutor)
	if err != nil {
		return err
	}
	getIDofUsers := func(users db.UserArray) []string {
		l := []string{}
		for _, user := range users {
			l = append(l, user.ID)
		}
		return l
	}
	err = lua.FillGroupUserRelationshipToRedis(groupID, getIDofUsers(users), cmdsExecutor)
	if err != nil {
		return err
	}
	return nil
}

type ScriptExecutor func(script *common.Script, keysAndArgs ...interface{}) (interface{}, error)
