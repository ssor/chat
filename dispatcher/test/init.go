package tests

import (
	"xsbPro/chat/dispatcher/dispatcher"
	"xsbPro/chat/dispatcher/resource"

	"github.com/ssor/config"
)

var (
	configContent = `{
    "listeningPort": "8092",
    "mode": "debug",
    "redisHost": ":6379",
    "mongoHost": "127.0.0.1",
    "dbName": "xsb_test",
}`

	conf config.IConfigInfo
)

func init() {
	var err error
	conf, err = config.ParseConfig([]byte(configContent))
	if err != nil {
		panic("配置文件加载错误: " + err.Error())
	}
	resource.Init(conf)

	dispatcher.ClearHistoryData(func() error {
		_, err := resource.RedisInstance.RedisDo("FLUSHDB")
		return err
	})

	// if lua.Lua_scripts == nil {
	// 	lua.Lua_scripts = lua.NewLuaScriptSet()
	// }
	// lua.InitLuaScripts(lua.Lua_scripts)
}
