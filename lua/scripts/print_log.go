package scripts

var PrintLog = `
		local function print_log(log)
			redis.pcall("echo", log)
		end
`
