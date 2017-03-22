package scripts

var FillNewGroup = `
		local function fill_new_group(key_of_group, args)
			local group_info = res2table(args)
			for key, value in pairs(group_info) do
				redis.call("hset", key_of_group, key, value)
			end
		redis.call("hset", key_of_group, "status", 1)	
		end
	`
