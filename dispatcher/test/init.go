package tests

import (
	dispatcher "xsbPro/chatDispatcher/dispatcher"
	"xsbPro/chatDispatcher/lua"
	"xsbPro/chatDispatcher/resource"

	"xsbPro/common"
)

var (
	config_content = `{
    "listeningPort": "8092",
    "mode": "debug",
    "redisHost": ":6379",
    "mongoHost": "127.0.0.1",
    "dbName": "xsb_test",
}`

	conf common.IConfigInfo
)

func init() {
	var err error
	conf, err = common.ParseConfig([]byte(config_content))
	if err != nil {
		panic("配置文件加载错误: " + err.Error())
	}
	if resource.Redis_instance == nil || resource.Mongo_pool == nil {
		resource.Init(conf)

		// InitRedis(conf)
		dispatcher.ClearHistoryData(func() error {
			_, err := resource.Redis_instance.RedisDo("FLUSHDB")
			return err
		})
	}

	if lua.Lua_scripts == nil {
		lua.Lua_scripts = lua.NewLuaScriptSet()
	}
	lua.InitLuaScripts(lua.Lua_scripts)
}
