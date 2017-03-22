package scripts

var GetNodeByGroup = `
		local function get_node_by_group(group_key)
			local node_key = redis.call("hget", group_key, "node")
			if (node_key) and (#node_key > 0) then
				local wan = redis.call("hget", node_key, "wan")
				return wan
			else
				return ""
			end
		end
	`
