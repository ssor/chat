package scripts

import (
	"fmt"
)

var GetUsersInGroup = fmt.Sprintf(`
local function get_users_in_group(group)
	local key_format_group_user_relationship = "%s"
	local key_format_userinfo = "%s"
	local users = {}
	local res_group_members = redis.call("SMEMBERS", string.format(key_format_group_user_relationship, group))
	if res_group_members then
		for _, user_id in ipairs(res_group_members) do
			local res_user = redis.call("HGETALL",  string.format(key_format_userinfo, user_id))
			table.insert(users, res2table(res_user))	
		end
	end

	return users
end
`, keyFormatGroupUserRelation, keyFormatUser)
