package scripts

var GetAllGroups = `
		local function get_all_groups()
			local groups = {}	
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				local res_group = redis.call("hgetall", key_of_group)
				table.insert(groups, res2table(res_group))
			end
			
			return groups
		end
	`
