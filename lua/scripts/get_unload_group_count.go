package scripts

var GetUnloadGroupCount = `
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
	`
