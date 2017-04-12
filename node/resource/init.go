package resource

import (
	"github.com/ssor/chat/redis"
	"github.com/ssor/config"
)

var (
	RedisInstance *redis.Instance
)

func init() {

}

// Init init resource
func Init(conf config.IConfigInfo) {
	RedisInstance = redis.NewInstance(conf.Get("redisHost").(string))
	if RedisInstance == nil {
		panic("init redis instance failed")
	}
}
