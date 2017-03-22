package scripts

var UpdateNodeCapability = `
		local function update_node_capability(node_key, cap)
			local exists = redis.call("EXISTS", node_key)
			if exists == 0 then
				return "ERROR"
			end

			redis.call("hset", node_key, "max", cap)
			return "OK"
		end
	`
