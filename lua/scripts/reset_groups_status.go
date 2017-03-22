package scripts

var ResetGroupsStatus = `
		local function reset_groups_status()
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				redis.call("hset", key_of_group, "status", 0)
			end
			return "OK"
		end
	`
