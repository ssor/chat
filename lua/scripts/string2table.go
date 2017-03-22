package scripts

var StringToTable = `
		local function string_to_table(s, sep)
			local tb = {}
			for w in string.gmatch(s,"([^"..sep.."]+)") do 
				table.insert(tb,w) 
			end
			return tb
		end
`
