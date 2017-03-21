package dispatcher

import (
	"fmt"
	"time"
	"xsbPro/chatDispatcher/lua"
	"xsbPro/chatDispatcher/resource"
	"xsbPro/common"
	"xsbPro/log"
	db "xsbPro/xsbdb"

	"github.com/parnurzeal/gorequest"
	"github.com/ssor/config"
	"github.com/ssor/redigo/redis"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	conf config.IConfigInfo

	node_alive_check_interval = 30 * time.Second
)

func init() {
}

func Init(_conf config.IConfigInfo, scriptExecutor ScriptExecutor, redisDo RedisDo) {
	conf = _conf

	prepareDataInRedis(_conf)

	go func() {
		tickerCheckNodeState := time.NewTicker(node_alive_check_interval)
		for {
			select {
			case <-tickerCheckNodeState.C:
				f := func(state, lan string) error {
					if state == node_state_dead {
						return RemoveNode(lan, scriptExecutor)
					}
					return nil
				}
				CheckNodeState(redisDo, f)
			}
		}
	}()
	// //remind all node to refresh data
	// f := func(state, lan string) error {
	// 	if state == node_state_alive {
	// 		log.InfoF("remind node to refresh data")
	// 		_, _, errs := gorequest.New().Get(getRefreshAllDataUrl(lan)).End()
	// 		if errs != nil && len(errs) > 0 {
	// 			log.InfoF("remind node to refresh all data error:%s", errs[0])
	// 			return fmt.Errorf("remind node to refresh all data error:%s", errs[0])
	// 		}
	// 	}
	// 	return nil
	// }
	// CheckNodeState(redisDo, f)
	go NotifyNodeDataRefresh(Datarefresh_refresh_all_data, "", redisDo)
}

func NotifyNodeDataRefresh(opt, para string, redisDo RedisDo) error {
	//remind all node to refresh data
	f := func(state, lan string) error {
		if state == node_state_alive {
			log.InfoF("remind node to refresh data")
			_, _, errs := gorequest.New().Get(getNodeDataRefreshDataUrl(lan, opt, para)).End()
			if errs != nil && len(errs) > 0 {
				log.InfoF("remind node to refresh data error:%s", errs[0])
				return fmt.Errorf("remind node to refresh data error:%s", errs[0])
			}
		}
		return nil
	}
	CheckNodeState(redisDo, f)
	return nil
}

func ClearHistoryData(cleaner func() error) {
	err := cleaner()
	if err != nil {
		panic(err)
	}
}

//整理
func prepareDataInRedis(_conf config.IConfigInfo) {
	// log.Info("prepareDataInRedis --->>>")
	session, err := resource.Mongo_pool.GetSession()
	if err != nil {
		panic("mongo conn err: " + err.Error())
	}
	defer resource.Mongo_pool.ReturnSession(session, err)
	err = FillDataToRedisFromMongo(session, _conf.Get("dbName").(string), resource.Redis_instance.RedisDoMulti, resource.Redis_instance.DoScript)
	if err != nil {
		panic(err)
	}
}

//要将 group 置为 非删除状态
func FillNewGroupToRedis(group *db.Group, scriptExecutor ScriptExecutor) error {
	args := redis.Args{}.Add(lua.Get_group_key(group.ID)).AddFlat(group)
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_fill_new_group], args...)
	if err != nil {
		log.SysF("FillNewGroupToRedis error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("FillNewGroupToRedis failed")
	}
	return nil
}

//将 redis 中的 group 置为删除状态
func ResetGroupsStatusInRedis(scriptExecutor ScriptExecutor) error {
	args := redis.Args{}
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_reset_groups_status], args...)
	if err != nil {
		log.SysF("ResetGroupsStatusInRedis error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("ResetGroupsStatusInRedis failed")
	}
	return nil
}

//由于 group 信息发生变化,对 node 的承载进行微调,尤其原来分配到 node 上的 group 已经被删除的情况下
func AdjustNodeDispatch(scriptExecutor ScriptExecutor) error {
	args := redis.Args{}
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_adjust_node_dispatch], args...)
	if err != nil {
		log.SysF("AdjustNodeDispatch error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("AdjustNodeDispatch failed")
	}
	return nil
}

// func FillGroupsToRedis(groups []*db.Group, cmdsExecutor func(*common.RedisCommands) error) error {
func FillGroupsToRedis(groups []*db.Group, scriptExecutor ScriptExecutor) error {

	for _, group := range groups {
		err := FillNewGroupToRedis(group, scriptExecutor)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveUsersFromRedis(users []string, cmdsExecutor func(*common.RedisCommands) error) error {

	cmds := common.NewRedisCommands(true)

	for _, user := range users {
		cmds.Add("DEL", redis.Args{}.Add(lua.Get_user_key(user)))
	}

	err := cmdsExecutor(cmds)
	return err
}

//clear users info in redis
func ClearUsersInRedis(scriptExecutor ScriptExecutor) error {
	args := redis.Args{}
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_clear_users], args...)
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

func FillUsersToRedis(users db.UserArray, cmdsExecutor func(*common.RedisCommands) error) error {

	cmds := common.NewRedisCommands(true)

	for _, user := range users {
		cmds.Add("HMSET", redis.Args{}.Add(lua.Get_user_key(user.ID)).AddFlat(user)...)
	}

	err := cmdsExecutor(cmds)
	return err
}

func UpdateUsersOfGroup(session *mgo.Session, db_name string, group string, cmdsExecutor func(*common.RedisCommands) error) error {
	//支部中用户信息数据重新填充
	users, err := GetUsersFromDB(session, db_name, bson.M{"group": group})
	if err != nil {
		return err
	}
	get_id_of_users := func(users db.UserArray) []string {
		l := []string{}
		for _, user := range users {
			l = append(l, user.ID)
		}
		return l
	}
	err = FillGroupUserRelationshipToRedis(group, get_id_of_users(users), cmdsExecutor)
	if err != nil {
		return err
	}
	return nil
}

func FillGroupUserRelationshipToRedis(group string, users []string, cmdsExecutor func(*common.RedisCommands) error) error {

	cmds := common.NewRedisCommands(true)

	key := lua.Get_group_user_relation_key(group)
	cmds.Add("DEL", redis.Args{}.Add(key))

	args := redis.Args{}.Add(key)
	for _, user := range users {
		args = args.AddFlat(user)
	}
	cmds.Add("SADD", args...)
	err := cmdsExecutor(cmds)
	return err
}

//每次载入数据,将 redis 中数据与 mongo 数据进行比对,完成后,检查分配信息,调整 node的实际承载量,需要的话,进行新的分配
func FillDataToRedisFromMongo(session *mgo.Session, db_name string, cmdsExecutor func(*common.RedisCommands) error, scriptExecutor ScriptExecutor) error {
	groups, err := GetGroupsFromDB(session, db_name, nil)
	if err != nil {
		return err
	}

	err = ResetGroupsStatusInRedis(scriptExecutor)
	if err != nil {
		return err
	}

	err = FillGroupsToRedis(groups, scriptExecutor)
	if err != nil {
		return err
	}

	err = AdjustNodeDispatch(scriptExecutor)
	if err != nil {
		return err
	}

	err = ClearUsersInRedis(scriptExecutor)
	if err != nil {
		return err
	}
	//支部中用户信息数据重新填充
	users, err := GetUsersFromDB(session, db_name, nil)
	if err != nil {
		return err
	}
	err = FillUsersToRedis(users, cmdsExecutor)
	if err != nil {
		return err
	}

	get_group_users := func(group *db.Group, users db.UserArray) db.UserArray {
		users_in_group := db.UserArray{}
		for _, user := range users {
			if user.Group == group.ID {
				users_in_group = append(users_in_group, user)
			}
		}
		return users_in_group
	}
	get_id_of_users := func(users db.UserArray) []string {
		l := []string{}
		for _, user := range users {
			l = append(l, user.ID)
		}
		return l
	}
	for _, group := range groups {
		users_in_group := get_group_users(group, users)
		// log.InfoF("group %s has %d users", group.ID, len(users_in_group))
		err = FillGroupUserRelationshipToRedis(group.ID, get_id_of_users(users_in_group), cmdsExecutor)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddUsers(session *mgo.Session, db_name string, query interface{}, cmdsExecutor func(*common.RedisCommands) error) error {
	users, err := GetUsersFromDB(session, db_name, query)
	if err != nil {
		return err
	}
	err = FillUsersToRedis(users, cmdsExecutor)
	if err != nil {
		return err
	}
	return nil
}

func GetUsersFromDB(session *mgo.Session, db_name string, query interface{}) (db.UserArray, error) {

	if session == nil {
		return nil, fmt.Errorf("db session should not be nil")
	}

	var users_array db.UserArray
	err := session.DB(db_name).C(common.Collection_userinfo).Find(query).All(&users_array)
	if err != nil {
		log.SysF("getUsersFromDB error: %s", err)
		return nil, err
	}

	return users_array, nil
}

func GetGroupsFromDB(session *mgo.Session, db_name string, query interface{}) ([]*db.Group, error) {

	if session == nil {
		return nil, fmt.Errorf("db session should not be nil")
	}

	var groups []*db.Group
	err := session.DB(db_name).C(common.Collection_group).Find(query).All(&groups)
	if err != nil {
		log.SysF("GetGroupsFromDB error: %s", err)
		return nil, err
	}

	return groups, nil
}

func AddNewGroup(session *mgo.Session, db_name, group_id string, cmdsExecutor func(*common.RedisCommands) error, scriptExecutor ScriptExecutor) error {
	groups, err := GetGroupsFromDB(session, db_name, bson.M{"_id": group_id})
	if err != nil {
		log.SysF("get group from db err: %s", err)
		return err
	}
	if len(groups) > 0 {
		err = FillGroupsToRedis(groups, scriptExecutor)
		if err != nil {
			log.SysF("FillGroupsToRedis  err: %s", err)
			return err
		}
	}
	//支部中用户信息数据重新填充
	users, err := GetUsersFromDB(session, db_name, bson.M{"group": group_id})
	if err != nil {
		return err
	}
	err = FillUsersToRedis(users, cmdsExecutor)
	if err != nil {
		return err
	}
	get_id_of_users := func(users db.UserArray) []string {
		l := []string{}
		for _, user := range users {
			l = append(l, user.ID)
		}
		return l
	}
	err = FillGroupUserRelationshipToRedis(group_id, get_id_of_users(users), cmdsExecutor)
	if err != nil {
		return err
	}
	return nil
}

type ScriptExecutor func(script *common.Script, keysAndArgs ...interface{}) (interface{}, error)
