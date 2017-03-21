package resource

import (
	"xsbPro/common"
)

var (
	RedisInstance *common.Redis_Instance
)

func init() {

}

// Init init resource
func Init(conf common.IConfigInfo) {
	RedisInstance = common.InitRedisInstance(conf.GetRedisHost())
	if RedisInstance == nil {
		panic("init redis instance failed")
	}
}
