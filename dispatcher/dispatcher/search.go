package dispatcher

import (
	"encoding/json"
	"strings"

	"github.com/ssor/chat/lua"
	"github.com/ssor/log"
)

type SearchResult struct {
	GroupID   string `json:"id" redis:"id"`
	GroupName string `json:"name" redis:"name"`
	Node      string `json:"node" redis:"node"`
}

type SearchResultArray []*SearchResult

func SearchGroupName(group string, scriptExecutor ScriptExecutor) (SearchResultArray, error) {

	// args := redis.Args{}
	// res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_all_groups], args...)
	res, err := lua.GetAllGroups(lua.ScriptExecutor(scriptExecutor))
	if err != nil {
		log.SysF("GetAllGroups error: %s", err)
		return nil, err
	}
	results := SearchResultArray{}
	var allResults SearchResultArray
	err = json.Unmarshal(res.([]uint8), &allResults)
	if err != nil {
		log.SysF("SearchGroupName error: %s", err)
		log.InfoF("-> %s", string(res.([]uint8)))
		return nil, err
	}
	log.TraceF("%s", string(res.([]uint8)))
	log.TraceF("got %d groups from db", len(allResults))
	if len(allResults) > 0 {
		for _, result := range allResults {
			// log.TraceF("name: %s ", result.GroupName)
			if strings.Contains(result.GroupName, group) {
				result.Node = strings.Replace(result.Node, "node->", "", 1)
				results = append(results, result)
			}
		}
	}
	return results, nil
}
