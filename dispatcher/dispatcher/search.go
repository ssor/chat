package dispatcher

import (
	"xsbPro/chatDispatcher/lua"
	"xsbPro/log"

	"encoding/json"

	"strings"

	"github.com/ssor/redigo/redis"
)

type SearchResult struct {
	GroupID   string `json:"id" redis:"id"`
	GroupName string `json:"name" redis:"name"`
	Node      string `json:"node" redis:"node"`
}

type SearchResultArray []*SearchResult

func SearchGroupName(group string, scriptExecutor ScriptExecutor) (SearchResultArray, error) {

	args := redis.Args{}
	res, err := scriptExecutor(lua.Lua_scripts.Scripts[lua.Lua_script_get_all_groups], args...)
	if err != nil {
		log.SysF("GetAllGroups error: %s", err)
		return nil, err
	}
	results := SearchResultArray{}
	var all_results SearchResultArray
	err = json.Unmarshal(res.([]uint8), &all_results)
	if err != nil {
		log.SysF("SearchGroupName error: %s", err)
		log.InfoF("-> %s", string(res.([]uint8)))
		return nil, err
	}
	log.TraceF("%s", string(res.([]uint8)))
	log.TraceF("got %d groups from db", len(all_results))
	if len(all_results) > 0 {
		for _, result := range all_results {
			// log.TraceF("name: %s ", result.GroupName)
			if strings.Contains(result.GroupName, group) {
				result.Node = strings.Replace(result.Node, "node->", "", 1)
				results = append(results, result)
			}
		}
	}
	return results, nil
}
