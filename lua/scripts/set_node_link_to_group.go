package scripts

var SetNodeLinkToGroup = `
		local function set_node_link_to_group(node_key, group_key)
			local exists = redis.call("EXISTS", node_key)
			if exists == 0 then
				return "ERROR"
			end
			local exists = redis.call("EXISTS", group_key)
			if exists == 0 then
				return "ERROR"
			end
			redis.call("hset", group_key, "node", node_key)	
			redis.call("HINCRBY", node_key, "current", 1)	
			return "OK"	
		end
	`
