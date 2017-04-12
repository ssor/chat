package lua

import (
	"fmt"

	"github.com/ssor/chat/lua/scripts"
)

var (
	luaScripts *luaScriptSet
)

func init() {
	if luaScripts == nil {
		luaScripts = newLuaScriptSet()
		initLuaScripts(luaScripts)
	}
}

const (
	luaScriptGetUsersInGroup      = "get_users_in_group"
	luaScriptUpdateNodeCapability = "update_node_capability"
	luaScriptGetUnloadGroupCount  = "get_unload_group_count"
	luaScriptGetNodeInfo          = "get_nodeinfo"
	luaScriptFillNewGroup         = "fill_new_group"
	luaScriptRemoveNode           = "remove_node"
	luaScriptRemoveGroup          = "remove_group"
	luaScriptGetNodeByGroup       = "get_node_by_group"
	luaScriptGetAllNodes          = "get_all_nodes"
	luaScriptGetGroupsOnNode      = "get_groups_on_node"
	luaScriptGetGroupUsersOnNode  = "get_group_users_on_node"
	luaScriptResetGroupsStatus    = "reset_groups_status"
	luaScriptAdjustNodeDispatch   = "adjust_node_dispatch"
	luaScriptClearUsers           = "clear_users"
	luaScriptGetAllGroups         = "get_all_groups"
	luaScriptSearchGroupName      = "search_group_name"
)

func initLuaScripts(set *luaScriptSet) {
	set.Add(luaScriptGetAllGroups, 0, fmt.Sprintf(`
	 	%s -- function res2table import
	 	%s -- function get_all_groups import

		local groups = get_all_groups()
		if #groups <= 0 then
			return "[]"
		else
			return cjson.encode(groups)
		end
	`, scripts.Res2table, scripts.GetAllGroups))
	// set.Add(luaScriptSearchGroupName, common.NewScript(1, fmt.Sprintf(`
	// 	%s -- function print_log import
	// 	%s -- function res2table import
	// 	%s -- function search_group_name import

	// 	local groups = search_group_name(KEYS[1])
	// 	if #groups <= 0 then
	// 		return "[]"
	// 	else
	// 		return cjson.encode(groups)
	// 	end
	// `,scripts.PrintLog , lua_func_res2table, lua_func_search_group_name)))

	set.Add(luaScriptClearUsers, 0, fmt.Sprintf(`
		%s -- function clear_users import
	
		return clear_users()
	`, scripts.ClearUsers))

	set.Add(luaScriptResetGroupsStatus, 0, fmt.Sprintf(`
		%s -- function reset_groups_status import
	
		return reset_groups_status()
	`, scripts.ResetGroupsStatus))

	set.Add(luaScriptAdjustNodeDispatch, 0, fmt.Sprintf(`
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
	`, scripts.PrintLog, scripts.Res2table, scripts.GetNodeInfo, scripts.FindNodeNotFull,
		scripts.GetAllNodes, scripts.GetGroupsUnlinked, scripts.GetGroupsOnNode,
		scripts.SetNodeLinkToGroup, scripts.AdjustNodeDispatch))

	set.Add(luaScriptGetGroupsOnNode, 1, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function get_groups_on_node import

		local groups = get_groups_on_node(KEYS[1]) 
		if #groups <= 0 then
			return "[]"
		else
			return cjson.encode(groups)
		end	
	`, scripts.Res2table, scripts.GetGroupsOnNode))

	set.Add(luaScriptRemoveGroup, 1, fmt.Sprintf(`
		%s -- function set_node_link_to_group import
		%s -- function remove_group import
		return remove_group(KEYS[1])
	`, scripts.SetNodeLinkToGroup, scripts.RemoveGroup))

	set.Add(luaScriptGetNodeByGroup, 1, fmt.Sprintf(`
		%s -- function get_node_by_group import
		
		return get_node_by_group(KEYS[1])
	`, scripts.GetNodeByGroup))

	set.Add(luaScriptRemoveNode, 1, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function get_nodeinfo import
		%s -- function find_node_not_full import
		%s -- function set_node_link_to_group import
		%s -- function remove_node import

		return remove_node(KEYS[1])
	`, scripts.Res2table, scripts.GetNodeInfo, scripts.FindNodeNotFull,
		scripts.SetNodeLinkToGroup, scripts.RemoveNode))

	set.Add(luaScriptFillNewGroup, 1, fmt.Sprintf(`
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
	`, scripts.Res2table, scripts.FillNewGroup, scripts.GetNodeInfo, scripts.FindNodeNotFull,
		scripts.SetNodeLinkToGroup))

	set.Add(luaScriptGetNodeInfo, 1, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function get_nodeinfo import

		return cjson.encode(get_nodeinfo(KEYS[1]))
	`, scripts.Res2table, scripts.GetNodeInfo))

	set.Add(luaScriptGetUsersInGroup, 0, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function string_to_table import
		%s -- function print_log import
		%s -- function table_length import
		%s -- function get_users_in_group import
		
		local group_id = ARGV[1]
		local users = get_users_in_group(group_id)
		
		return cjson.encode(users)
	`, scripts.Res2table, scripts.StringToTable, scripts.PrintLog,
		scripts.TableLength, scripts.GetUsersInGroup,
	))

	set.Add(luaScriptUpdateNodeCapability, 1, fmt.Sprintf(`
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
	`, scripts.PrintLog, scripts.Res2table, scripts.UpdateNodeCapability, scripts.GetNodeInfo,
		scripts.GetGroupsUnlinked, scripts.SetNodeLinkToGroup))

	set.Add(luaScriptGetUnloadGroupCount, 0, fmt.Sprintf(`
		%s -- function get_unload_group_count import

		return get_unload_group_count()	
	`, scripts.GetUnloadGroupCount))

	set.Add(luaScriptGetAllNodes, 0, fmt.Sprintf(`
		%s -- function res2table import
		%s -- function get_nodeinfo import
		%s -- function get_all_nodes import
		local nodes = get_all_nodes()
		return cjson.encode(nodes)
	`, scripts.Res2table, scripts.GetNodeInfo, scripts.GetAllNodes))

	set.Add(luaScriptGetGroupUsersOnNode, 0, fmt.Sprintf(`
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
	`, scripts.Res2table, scripts.StringToTable, scripts.PrintLog,
		scripts.TableLength, scripts.GetUsersInGroup, scripts.GetGroupUsersOnNode))
}
