package scripts

var GetGroupUsersOnNode = `
		local function get_group_users_on_node(group_key, node_key, group)
			local users = {}
			local res_node = redis.call("hget", group_key, "node")
			if (res_node) and (res_node == node_key) then
				return "OK", get_users_in_group(group)	
			end

			return "FAILED", users
		end
	`
