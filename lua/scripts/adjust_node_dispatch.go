package scripts

var AdjustNodeDispatch = `
		local function adjust_node_dispatch()
			--remove all groups that do not exist
			local keys_of_group = redis.call("keys", "group:*")
			for index, key_of_group in ipairs(keys_of_group) do 
				local res = redis.call("hget", key_of_group, "status")
				if (res == "0") then
					redis.call("del", key_of_group)
				end 
			end

			--count groups on each node
			local nodes = get_all_nodes()
			if (#nodes > 0) then
				for index, node in ipairs(nodes) do
					local node_key = "node->" .. node.lan
					local groups = get_groups_on_node(node_key)
					redis.call("hset", node_key, "current", #groups)
				end
			end
			
			--redispatch group on node
			local groups = get_groups_unlinked(100000000)
			if (#groups > 0) then
				print_log(#groups .. " unlinked group found ----<<<<")
				for index, group in ipairs(groups) do
					local node_not_full_key = find_node_not_full()
						print_log("find_node_not_full -> " .. node_not_full_key .. " --<<<<")
					if node_not_full_key ~= "" then
						local res = set_node_link_to_group(node_not_full_key, group)
						print_log("set_node_link_to_group res -> " .. res .. " ---<<<")
						if res ~= "OK" then
							return res
						end
					else	
						return "OK"
					end
				end
			end

			return "OK"
		end
	`
