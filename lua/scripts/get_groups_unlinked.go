package scripts

var GetGroupsUnlinked = `
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
	`
