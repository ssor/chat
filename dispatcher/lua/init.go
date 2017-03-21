package lua

import (
	"fmt"
	"xsbPro/common"
	"xsbPro/log"
	db "xsbPro/xsbdb"

	"github.com/ssor/redigo/redis"

	"encoding/json"

	"strconv"
	"time"
)

var (
	Lua_scripts *LuaScriptSet
)

func init() {
	Lua_scripts = NewLuaScriptSet()
	InitLuaScripts(Lua_scripts)
}

const (
	Lua_script_get_users_in_group      = "get_users_in_group"
	Lua_script_update_node_capability  = "update_node_capability"
	Lua_script_get_unload_group_count  = "get_unload_group_count"
	Lua_script_get_nodeinfo            = "get_nodeinfo"
	Lua_script_fill_new_group          = "fill_new_group"
	Lua_script_remove_node             = "remove_node"
	Lua_script_remove_group            = "remove_group"
	Lua_script_get_node_by_group       = "get_node_by_group"
	Lua_script_get_all_nodes           = "get_all_nodes"
	Lua_script_get_groups_on_node      = "get_groups_on_node"
	Lua_script_get_group_users_on_node = "get_group_users_on_node"
	Lua_script_reset_groups_status     = "reset_groups_status"
	Lua_script_adjust_node_dispatch    = "adjust_node_dispatch"
	Lua_script_clear_users             = "clear_users"
	Lua_script_get_all_groups          = "get_all_groups"
	Lua_script_search_group_name       = "search_group_name"
)

func InitLuaScripts(set *LuaScriptSet) {
	set.Add(Lua_script_get_all_groups, common.NewScript(0, fmt.Sprintf(`
	 	%s -- function res2table import
	 	%s -- function get_all_groups import

		local groups = get_all_groups()
		if #groups <= 0 then
			return "[]"
		else
			return cjson.encode(groups)
		end
	`, lua_func_res2table, lua_func_get_all_groups)))
	// set.Add(Lua_script_search_group_name, common.NewScript(1, fmt.Sprintf(`
	// 	%s -- function print_log import
	// 	%s -- function res2table import
	// 	%s -- function search_group_name import

	// 	local groups = search_group_name(KEYS[1])
	// 	if #groups <= 0 then
	// 		return "[]"
	// 	else
	// 		return cjson.encode(groups)
	// 	end
	// `, lua_func_log, lua_func_res2table, lua_func_search_group_name)))

	set.Add(Lua_script_clear_users, common.NewScript(0, fmt.Sprintf(`
		%s -- function clear_users import
	
		return clear_users()
	`, lua_func_clear_users)))

	set.Add(Lua_script_reset_groups_status, common.NewScript(0, fmt.Sprintf(`
		%s -- function reset_groups_status import
	
		return reset_groups_status()
	`, lua_func_reset_groups_status)))

	set.Add(Lua_script_adjust_node_dispatch, common.NewScript(0, fmt.Sprintf(`
		%s -- function print_log import
		%s -- function res2table import
		%s -- function get_nodeinfo import
		%s -- function find_node_not_full import	
		%s -- function get_all_nodes import	
		%s -- function get_groups_unlinked import	
		%s -- function get_groups_on_node import
		%s -- function set_node_link_to_group import
		%s -- function adjust_node_dispatch import

		return adjust_node_dispatch()
	`, lua_func_log, lua_func_res2table, lua_func_get_nodeinfo, lua_func_find_node_not_full,
		lua_func_get_all_nodes, lua_func_get_groups_unlinked, lua_func_get_groups_on_node,
		lua_func_set_node_link_to_group, lua_func_adjust_node_dispatch)))

	set.Add(Lua_script_get_groups_on_node, common.NewScript(1, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function get_groups_on_node import

		local groups = get_groups_on_node(KEYS[1]) 
		if #groups <= 0 then
			return "[]"
		else
			return cjson.encode(groups)
		end	
	`, lua_func_res2table, lua_func_get_groups_on_node)))

	set.Add(Lua_script_remove_group, common.NewScript(1, fmt.Sprintf(`
		%s -- function set_node_link_to_group import
		%s -- function remove_group import
		return remove_group(KEYS[1])
	`, lua_func_set_node_link_to_group, lua_func_remove_group)))

	set.Add(Lua_script_get_node_by_group, common.NewScript(1, fmt.Sprintf(`
		%s -- function get_node_by_group import
		
		return get_node_by_group(KEYS[1])
	`, lua_func_get_node_by_group)))

	set.Add(Lua_script_remove_node, common.NewScript(1, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function get_nodeinfo import
		%s -- function find_node_not_full import
		%s -- function set_node_link_to_group import
		%s -- function remove_node import

		return remove_node(KEYS[1])
	`, lua_func_res2table, lua_func_get_nodeinfo, lua_func_find_node_not_full,
		lua_func_set_node_link_to_group, lua_func_remove_node)))

	set.Add(Lua_script_fill_new_group, common.NewScript(1, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function fill_new_group import
		%s -- function get_nodeinfo import
		%s -- function find_node_not_full import
		%s -- function set_node_link_to_group import

		local group_key = KEYS[1]	
		fill_new_group(group_key, ARGV)
		local not_full_node_key = find_node_not_full()	
		if not_full_node_key ~= "" then
			return set_node_link_to_group(not_full_node_key, group_key)	
		end	
		return "OK"	
	`, lua_func_res2table, lua_func_fill_new_group, lua_func_get_nodeinfo, lua_func_find_node_not_full,
		lua_func_set_node_link_to_group)))

	set.Add(Lua_script_get_nodeinfo, common.NewScript(1, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function get_nodeinfo import

		return cjson.encode(get_nodeinfo(KEYS[1]))
	`, lua_func_res2table, lua_func_get_nodeinfo)))

	set.Add(Lua_script_get_users_in_group, common.NewScript(0, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function string_to_table import
		%s -- function print_log import
		%s -- function table_length import
		%s -- function get_users_in_group import
		
		local group_id = ARGV[1]
		local users = get_users_in_group(group_id)
		
		return cjson.encode(users)
	`, lua_func_res2table, lua_func_string_to_table, lua_func_log,
		lua_func_table_length, lua_func_get_users_in_group,
	)))

	set.Add(Lua_script_update_node_capability, common.NewScript(1, fmt.Sprintf(`
		%s -- function print_log import
		%s -- function res2table import
		%s -- function update_node_capability import
		%s -- function get_nodeinfo import
		%s -- function get_groups_unlinked import
		%s -- function set_node_link_to_group import
		local node_key = KEYS[1]
		local cap = ARGV[1]

		if update_node_capability(node_key, cap) == "OK" then	
			local nodeinfo = get_nodeinfo(node_key)
			if nodeinfo.max > nodeinfo.current then
				local groups_unlinked = get_groups_unlinked(nodeinfo.max - nodeinfo.current)
				for index, group_key in ipairs(groups_unlinked) do	
					local res = set_node_link_to_group(node_key, group_key)
					if res ~= "OK" then
						return "ERROR"
					end
				end
			end
		else
			return "ERROR"
		end 
		return "OK"
	`, lua_func_log, lua_func_res2table, lua_func_update_node_capability, lua_func_get_nodeinfo,
		lua_func_get_groups_unlinked, lua_func_set_node_link_to_group)))

	set.Add(Lua_script_get_unload_group_count, common.NewScript(0, fmt.Sprintf(`
		%s -- function get_unload_group_count import

		return get_unload_group_count()	
	`, lua_func_get_unload_group_count)))

	set.Add(Lua_script_get_all_nodes, common.NewScript(0, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function get_nodeinfo import
		%s -- function get_all_nodes import
		local nodes = get_all_nodes()
		return cjson.encode(nodes)
	`, lua_func_res2table, lua_func_get_nodeinfo, lua_func_get_all_nodes)))

	set.Add(Lua_script_get_group_users_on_node, common.NewScript(0, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function string_to_table import
		%s -- function print_log import
		%s -- function table_length import	
		%s -- function get_users_in_group import
		%s -- function get_group_users_on_node import

		local group_key = ARGV[1]
		local node_key = ARGV[2]
		local group = ARGV[3]
		local result, users = get_group_users_on_node(group_key, node_key, group)
		if result == "OK" then
			return cjson.encode(users)
		else
			return "[]"	
		end	
	`, lua_func_res2table, lua_func_string_to_table, lua_func_log,
		lua_func_table_length, lua_func_get_users_in_group, lua_func_get_group_users_on_node)))
}

func RemoveGroupFromRedis(group string, scriptExecutor ScriptExecutor) error {
	args := redis.Args{}.Add(fmt.Sprintf(key_format_group, group))
	res, err := scriptExecutor(Lua_scripts.Scripts[Lua_script_remove_group], args...)
	if err != nil {
		log.SysF("RemoveGroupFromRedis error: %s", err)
		return err
	}
	if string(res.([]uint8)) != "OK" {
		return fmt.Errorf("RemoveGroupFromRedis failed")
	}
	return nil
}

type ScriptExecutor func(script *common.Script, keysAndArgs ...interface{}) (interface{}, error)

func GetGroupUsersFromCache(group string, scriptExecutor ScriptExecutor) (db.UserArray, error) {
	args := redis.Args{}.AddFlat(group)
	res, err := scriptExecutor(Lua_scripts.Scripts[Lua_script_get_users_in_group], args...)
	if err != nil {
		log.InfoF("GetGroupUsersFromCache DoScript err: %s", err)
		return nil, err
	}
	data := res.([]uint8)

	if string(data) == "{}" {
		return db.UserArray{}, nil
	}

	return unmarshalUsers(data), nil
	// type UserCache struct {
	// 	ID            string `bson:"_id" redis:"id"`
	// 	Password      string `bson:"password" redis:"password"`
	// 	Name          string `bson:"name" redis:"name"`
	// 	Phone         string `bson:"phone" redis:"phone"`
	// 	Email         string `bson:"email" redis:"email"`
	// 	Group         string `bson:"group" redis:"group"`
	// 	Index         string `bson:"index" redis:"index"`
	// 	Actived       string `bson:"actived" redis:"actived"`
	// 	Image         string `bson:"image" redis:"image"`
	// 	Tag           string `bson:"tag" redis:"tag"`
	// 	Chief         string `bson:"chief" redis:"chief"`
	// 	Gender        string `bson:"gender" redis:"gender"`
	// 	Department    string `bson:"department" redis:"department"`
	// 	UserType      string `bson:"userType" redis:"userType"`
	// 	ActiveCode    string `bson:"activeCode" redis:"activeCode"`
	// 	LastLoginTime string `bson:"lastLoginTime" redis:"-"`
	// 	LastLoginAddr string `bson:"lastLoginAddr" redis:"-"`
	// 	JoinPartyDate string `bson:"joinPartyDate" redis:"joinPartyDate"`
	// 	BirthDay      string `bson:"birthDay" redis:"birthDay"`
	// 	Title         string `bson:"title" redis:"title"`
	// 	Community     string `bson:"community" redis:"community"`
	// }
	// var users_cache []*UserCache
	// err = json.Unmarshal(res.([]uint8), &users_cache)
	// if err != nil {
	// 	log.Info(string(res.([]uint8)))
	// 	log.InfoF("GetGroupUsersFromCache  err: %s", err)
	// 	return nil, err
	// }

	// convert_user_cache_to_user := func(cache *UserCache) *db.User {
	// 	index, err := strconv.Atoi(cache.Index)
	// 	if err != nil {
	// 		index = 0
	// 	}
	// 	actived, err := strconv.Atoi(cache.Actived)
	// 	if err != nil {
	// 		actived = 0
	// 	}
	// 	gender, err := strconv.Atoi(cache.Gender)
	// 	if err != nil {
	// 		gender = 0
	// 	}
	// 	user_type, err := strconv.Atoi(cache.UserType)
	// 	if err != nil {
	// 		user_type = 0
	// 	}
	// 	chief, err := strconv.ParseBool(cache.Chief)
	// 	if err != nil {
	// 		chief = false
	// 	}
	// 	user := db.NewUser(cache.ID, cache.Phone, cache.Email, cache.Password, cache.Name, "",
	// 		cache.Group, cache.Image, cache.Department, cache.Tag, gender, index, chief, user_type, cache.ActiveCode)
	// 	user.Actived = actived
	// 	return user
	// }
	// users := db.UserArray{}
	// for _, cache := range users_cache {
	// 	users = append(users, convert_user_cache_to_user(cache))
	// }
	// return users, nil

}

func unmarshalUsers(data []byte) db.UserArray {

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
	var users_cache []*UserCache
	err := json.Unmarshal(data, &users_cache)
	if err != nil {
		log.Info(string(data))
		log.InfoF("unmarshalUsers  err: %s", err)
		return nil
	}

	convert_user_cache_to_user := func(cache *UserCache) *db.User {
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
		user_type, err := strconv.Atoi(cache.UserType)
		if err != nil {
			user_type = 0
		}
		chief, err := strconv.ParseBool(cache.Chief)
		if err != nil {
			chief = false
		}
		user := db.NewUser(cache.ID, cache.Phone, cache.Email, cache.Password, cache.Name, "",
			cache.Group, cache.Image, cache.Department, cache.Tag, gender, index, chief, user_type, cache.ActiveCode, time.Now().Format(time.RFC3339))
		user.Actived = actived
		return user
	}
	users := db.UserArray{}
	for _, cache := range users_cache {
		users = append(users, convert_user_cache_to_user(cache))
	}
	return users
}

func GetGroupsOnNode(lan string, scriptExecutor ScriptExecutor) ([]*db.Group, error) {
	node_key := Format_nodeinfo_key(lan)
	args := redis.Args{}.Add(node_key)
	res, err := scriptExecutor(Lua_scripts.Scripts[Lua_script_get_groups_on_node], args...)
	if err != nil {
		log.SysF("GetGroupsOnNode error: %s", err)
		return nil, err
	}
	var groups []*db.Group
	err = json.Unmarshal(res.([]uint8), &groups)
	if err != nil {
		log.InfoF("data: %s", string(res.([]uint8)))
		return nil, err
	}
	return groups, nil
}

//the group must be dispatched to specified node, or no data should be returned
func GetGroupUsersOnNode(group, node_lan string, scriptExecutor ScriptExecutor) (db.UserArray, error) {

	args := redis.Args{}.AddFlat(Get_group_key(group)).AddFlat(Format_nodeinfo_key(node_lan)).AddFlat(group)
	res, err := scriptExecutor(Lua_scripts.Scripts[Lua_script_get_group_users_on_node], args...)
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
