package scripts

var FindNodeNotFull = `
		local function find_node_not_full()
			local node_keys = redis.call("keys", "node->*")
			for index, node_key in ipairs(node_keys) do 
				local node = get_nodeinfo(node_key)
				if node.max == nil then
					node.max = 0
				end
				
				if tonumber(node.current) < tonumber(node.max) then
					return node_key
				end
			end
			return ""
		end
	`
