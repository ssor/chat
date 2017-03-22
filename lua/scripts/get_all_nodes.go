package scripts

var GetAllNodes = `
		local function get_all_nodes()
			local nodes = {}	
			local node_keys = redis.call("keys", "node->*")
			for index, node_key in ipairs(node_keys) do 
				local node = get_nodeinfo(node_key)
				table.insert(nodes, node)	
			end
			return nodes
		end
	`
