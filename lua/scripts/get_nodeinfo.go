package scripts

var GetNodeInfo = `
		local function get_nodeinfo(node_key)
			local res_node = redis.call("hgetall", node_key)
			local nodeinfo = res2table(res_node)
			return nodeinfo
		end
	`
