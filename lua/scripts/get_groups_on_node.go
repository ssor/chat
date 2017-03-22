package scripts

var GetGroupsOnNode = `
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
	`
