package scripts

var Res2table = `
		local function res2table(res)
			local tb = {}
			if res then
				local max = #res
				for index = 1,max,2 do
					tb[res[index]] = res[index+1]
				end
			end
			return tb
		end
`
