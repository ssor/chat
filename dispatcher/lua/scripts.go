package lua

import (
	"fmt"
	"xsbPro/common"
)

const (
	key_format_group_user_relation = "group->users:%s"
	key_format_group               = "group:%s"
	key_format_user                = "user:%s"
)

func Format_nodeinfo_key(lan string) string {
	return fmt.Sprintf(common.Hash_node, lan)
}

// func Format_group_node_map_key(group, lan, wan string) string {
// 	return fmt.Sprintf(common.String_group_node_map, group, lan, wan)
// }

func Get_group_key(group string) string {
	return fmt.Sprintf(key_format_group, group)
}

func Get_user_key(user string) string {
	return fmt.Sprintf(key_format_user, user)
}

func Get_group_user_relation_key(group string) string {
	return fmt.Sprintf(key_format_group_user_relation, group)
}

var (
	lua_func_res2table = `
		local function res2table(res)
			local tb = {}
			if res then
				local max = #res
				for index = 1,max,2 do
					tb[res[index]] = res[index+1]
				end
			end
			return tb
		end
	`

	lua_func_string_to_table = `
		local function string_to_table(s, sep)
			local tb = {}
			for w in string.gmatch(s,"([^"..sep.."]+)") do 
				table.insert(tb,w) 
			end
			return tb
		end
	`

	lua_func_log = `
		local function print_log(log)
			redis.pcall("echo", log)
		end
	`
	lua_func_table_length = `
		local function table_length(T)
			local count = 0
			for _ in pairs(T) do count = count + 1 end
			return count
		end
	`
	lua_func_clear_users = fmt.Sprintf(`
		local function clear_users()
			local users = redis.call("KEYS", "user:*")
			if users then
				for _,user in ipairs(users) do
					redis.call("DEL", user)
				end
				--redis.call("DEL", unpack(users))
			end
			local groups = redis.call("KEYS", "group->users:*")
			if groups then
				for _, group in ipairs(groups) do
					redis.call("DEL", group)
				end
			end
			return "OK"
		end	
	`)

	lua_func_get_users_in_group = fmt.Sprintf(`
		local key_format_group_user_relationship = "%s"
		local key_format_userinfo = "%s"

		local function get_users_in_group(group)
			local users = {}
			local res_group_members = redis.call("SMEMBERS", string.format(key_format_group_user_relationship, group))
			if res_group_members then
				for _, user_id in ipairs(res_group_members) do
					local res_user = redis.call("HGETALL",  string.format(key_format_userinfo, user_id))
					table.insert(users, res2table(res_user))	
				end
			end

			return users
		end
	`, key_format_group_user_relation, key_format_user)

	lua_func_get_groups_unlinked = fmt.Sprintf(`
		-- return a list of group key
		local function get_groups_unlinked(max)
			local list = {}
			if max <= 0 then
				return list
			end
			local count = 0
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				local res_node_key = redis.call("hget", key_of_group, "node")	
				if  (not res_node_key) or #res_node_key <= 0 then
					table.insert(list, key_of_group)
					count = count + 1
					if count >= max then
						return list
					end
				end
			end
			return list 
		end
	`)

	lua_func_update_node_capability = fmt.Sprintf(`
		local function update_node_capability(node_key, cap)
			local exists = redis.call("EXISTS", node_key)
			if exists == 0 then
				return "ERROR"
			end

			redis.call("hset", node_key, "max", cap)
			return "OK"
		end
	`)

	lua_func_get_unload_group_count = fmt.Sprintf(`
		local function get_unload_group_count()
		local count = 0
		local keys_of_group = redis.call("keys", "group:*")
		for index, key_of_group in ipairs(keys_of_group) do 
			local res_node = redis.call("hget", key_of_group, "node")
			if (not res_node) or (#res_node <= 0) then
				count = count + 1
			end
		end
		return count
		end
	`)

	lua_func_get_nodeinfo = fmt.Sprintf(`
		local function get_nodeinfo(node_key)
			local res_node = redis.call("hgetall", node_key)
			local nodeinfo = res2table(res_node)
			return nodeinfo
		end
	`)

	lua_func_find_node_not_full = fmt.Sprintf(`
		local function find_node_not_full()
			local node_keys = redis.call("keys", "node->*")
			for index, node_key in ipairs(node_keys) do 
				local node = get_nodeinfo(node_key)
				if node.max == nil then
					node.max = 0
				end
				
				if tonumber(node.current) < tonumber(node.max) then
					return node_key
				end
			end
			return ""
		end
	`)

	lua_func_set_node_link_to_group = fmt.Sprintf(`
		local function set_node_link_to_group(node_key, group_key)
			local exists = redis.call("EXISTS", node_key)
			if exists == 0 then
				return "ERROR"
			end
			local exists = redis.call("EXISTS", group_key)
			if exists == 0 then
				return "ERROR"
			end
			redis.call("hset", group_key, "node", node_key)	
			redis.call("HINCRBY", node_key, "current", 1)	
			return "OK"	
		end
	`)

	lua_func_fill_new_group = fmt.Sprintf(`
		local function fill_new_group(key_of_group, args)
			local group_info = res2table(args)
			for key, value in pairs(group_info) do
				redis.call("hset", key_of_group, key, value)
			end
		redis.call("hset", key_of_group, "status", 1)	
		end
	`)

	lua_func_reset_groups_status = fmt.Sprintf(`
		local function reset_groups_status()
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				redis.call("hset", key_of_group, "status", 0)
			end
			return "OK"
		end
	`)

	lua_func_adjust_node_dispatch = fmt.Sprintf(`
		local function adjust_node_dispatch()
			--remove all groups that do not exist
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				local res = redis.call("hget", key_of_group, "status")
				if (res == "0") then
					redis.call("del", key_of_group)
				end 
			end

			--count groups on each node
			local nodes = get_all_nodes()
			if (#nodes > 0) then
				for index, node in ipairs(nodes) do
					local node_key = "node->" .. node.lan
					local groups = get_groups_on_node(node_key)
					redis.call("hset", node_key, "current", #groups)
				end
			end
			
			--redispatch group on node
			local groups = get_groups_unlinked(100000000)
			if (#groups > 0) then
				print_log(#groups .. " unlinked group found ----<<<<")
				for index, group in ipairs(groups) do
					local node_not_full_key = find_node_not_full()
						print_log("find_node_not_full -> " .. node_not_full_key .. " --<<<<")
					if node_not_full_key ~= "" then
						local res = set_node_link_to_group(node_not_full_key, group)
						print_log("set_node_link_to_group res -> " .. res .. " ---<<<")
						if res ~= "OK" then
							return res
						end
					else	
						return "OK"
					end
				end
			end

			return "OK"
		end
	`)

	lua_func_remove_node = fmt.Sprintf(`
		local function remove_node(key_of_node)
			local exists = redis.call("EXISTS", key_of_node)
			if exists == 0 then
				return "ERROR"
			end

			local keys_of_group = redis.call("keys", "group:*")
			local unlink_groups = {}
			for index, key_of_group in ipairs(keys_of_group) do 
				local res_node = redis.call("hget", key_of_group, "node")
				if (res_node) and (res_node == key_of_node) then
					redis.call("hset", key_of_group, "node", "")
					table.insert(unlink_groups, key_of_group)
				end
			end

			redis.call("DEL", key_of_node)

			for index, key_of_group in ipairs(unlink_groups) do 
				local not_full_node = find_node_not_full()
				if 	not_full_node == "" then
					return "OK"
				end
				local res = set_node_link_to_group(not_full_node, key_of_group)
				if res == "ERROR" then
					return "ERROR"
				end
			end
	
			return "OK"
		end
	`)

	lua_func_get_node_by_group = fmt.Sprintf(`
		local function get_node_by_group(group_key)
			local node_key = redis.call("hget", group_key, "node")
			if (node_key) and (#node_key > 0) then
				local wan = redis.call("hget", node_key, "wan")
				return wan
			else
				return ""
			end
		end
	`)

	lua_func_remove_group = fmt.Sprintf(`
		local function remove_group(group_key)
			local node_key = redis.call("hget", group_key, "node")
			if (node_key) and (#node_key > 0) then
				redis.call("HINCRBY", node_key, "current", -1)						
			end
			
			redis.call("del", group_key)			
			
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				local res_node = redis.call("hget", key_of_group, "node")
				if (not res_node) or (#res_node <= 0) then
					return set_node_link_to_group(node_key, key_of_group)
				end
			end

			return "OK"
		end
	`)

	lua_func_get_all_nodes = fmt.Sprintf(`
		local function get_all_nodes()
			local nodes = {}	
			local node_keys = redis.call("keys", "node->*")
			for index, node_key in ipairs(node_keys) do 
				local node = get_nodeinfo(node_key)
				table.insert(nodes, node)	
			end
			return nodes
		end
	`)

	lua_func_get_groups_on_node = fmt.Sprintf(`
		local function get_groups_on_node(node_key)
			local groups = {}	
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				local res_node = redis.call("hget", key_of_group, "node")
				if (res_node) and (res_node == node_key) then
					local res_group = redis.call("hgetall", key_of_group)
					table.insert(groups, res2table(res_group))
				end
			end
			
			return groups
		end
	`)

	lua_func_get_group_users_on_node = fmt.Sprintf(`
		local function get_group_users_on_node(group_key, node_key, group)
			local users = {}
			local res_node = redis.call("hget", group_key, "node")
			if (res_node) and (res_node == node_key) then
				return "OK", get_users_in_group(group)	
			end

			return "FAILED", users
		end
	`)

	lua_func_get_all_groups = fmt.Sprintf(`
		local function get_all_groups()
			local groups = {}	
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				local res_group = redis.call("hgetall", key_of_group)
				table.insert(groups, res2table(res_group))
			end
			
			return groups
		end
	`)

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
