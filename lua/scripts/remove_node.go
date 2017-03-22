package scripts

var RemoveNode = `
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
	`
