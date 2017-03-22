package scripts

var ClearUsers = `
		local function clear_users()
			local users = redis.call("KEYS", "user:*")
			if users then
				for _,user in ipairs(users) do
					redis.call("DEL", user)
				end
				--redis.call("DEL", unpack(users))
			end
			local groups = redis.call("KEYS", "group->users:*")
			if groups then
				for _, group in ipairs(groups) do
					redis.call("DEL", group)
				end
			end
			return "OK"
		end	
	`
