package resource

import (
	"xsbPro/common"

	"github.com/ssor/config"

	"github.com/ssor/mongopool"
)

var (
	//RedisInstance redis instance
	RedisInstance *common.Redis_Instance
	// MongoPool mongo instance
	MongoPool *mongo_pool.MongoSessionPool
)

// Init give a tool to init db resource
func Init(conf config.IConfigInfo) {
	initMongo(conf)
	initRedis(conf)
}

func initRedis(conf config.IConfigInfo) {
	if RedisInstance == nil {
		RedisInstance = common.InitRedisInstance(conf.Get("redisHost").(string))
	}
}

func initMongo(conf config.IConfigInfo) {
	if MongoPool == nil {
		MongoPool = mongo_pool.NewMongoSessionPool(conf.Get("mongoHost").(string), 3)
		MongoPool.Run()
	}
}
