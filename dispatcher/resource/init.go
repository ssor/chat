package resource

import (
	"xsbPro/common"

	"github.com/ssor/config"

	"github.com/ssor/mongopool"
)

var (
	Redis_instance *common.Redis_Instance
	Mongo_pool     *mongo_pool.MongoSessionPool
)

func Init(conf config.IConfigInfo) {
	initMongo(conf)
	initRedis(conf)
}

func initRedis(conf config.IConfigInfo) {
	Redis_instance = common.InitRedisInstance(conf.Get("redisHost").(string))
}

func initMongo(conf config.IConfigInfo) {
	Mongo_pool = mongo_pool.NewMongoSessionPool(conf.Get("mongoHost").(string), 3)
	Mongo_pool.Run()
}
