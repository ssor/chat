package scripts

var RemoveGroup = `
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
	`
