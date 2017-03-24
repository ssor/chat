package resource

import (
	"xsbPro/common"

	"github.com/ssor/config"
)

var (
	RedisInstance *common.Redis_Instance
)

func init() {

}

// Init init resource
func Init(conf config.IConfigInfo) {
	RedisInstance = common.InitRedisInstance(conf.Get("redisHost").(string))
	if RedisInstance == nil {
		panic("init redis instance failed")
	}
}
