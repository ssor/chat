package controller

import (
	"encoding/json"
	"sync"
	"time"

	nsq "github.com/nsqio/go-nsq"
	"github.com/ssor/chat/dispatcher/dispatcher"
	"github.com/ssor/chat/dispatcher/resource"
	"github.com/ssor/chat/lua"
	"github.com/ssor/chat/redis"
	"github.com/ssor/config"
	"github.com/ssor/log"
	"gopkg.in/mgo.v2/bson"
)

var (
	lookupdPollInterval = 15 * time.Second
	nsqChannel          = "chatdispatcher"
	conf                config.IConfigInfo

	waitForDbConnection = sync.Mutex{}
)

// Init init resources
func Init(c config.IConfigInfo) {
	conf = c

	resource.Init(conf)

	dispatcher.Init(conf, resource.RedisInstance.DoScript, resource.RedisInstance.RedisDo)

	//当支部发生变化时的处理
	//1. 新添加了支部    -> 添加新支部信息,更新支部人员信息,更新人员和支部关系,只需分配节点
	//3. 删除支部       -> 清理数据,通知节点
	//4. 支部人员信息更新 ->  更新支部人员信息,更新人员和支部关系,通知节点
	//5. 人员信息增加    -> 更新人员信息,更新人员和支部关系
	//6. 人员信息更新    -> 更新人员信息
	//7. 人员删除       -> 移除人员信息,更新人员和支部关系

	startNsqConsumer(conf.Get("nsqHost").(string), nsqTopicGroup, nsq.HandlerFunc(updateGroup))
	startNsqConsumer(conf.Get("nsqHost").(string), nsqTopicUser, nsq.HandlerFunc(updateUser))
	startNsqConsumer(conf.Get("nsqHost").(string), nsqTopicUsersOfGroupUpdate, nsq.HandlerFunc(updateUsersOfGroup))

}

func init() {

}

func updateUsersOfGroup(msg *nsq.Message) error {
	type InnerSyncMessage struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}
	var ins InnerSyncMessage
	err := json.Unmarshal(msg.Body, &ins)
	if err != nil {
		log.SysF("updateUsersOfGroup err: %s", err)
		return err
	}

	err = updateUsersOfGroupToDB(conf.Get("dbName").(string), ins.Data, resource.RedisInstance.RedisDoMulti)
	if err != nil {
		log.SysF("updateUsersOfGroup err: %s", err)
		return err
	}
	//notify node to remove group
	err = dispatcher.NotifyNodeDataRefresh(dispatcher.DataRefreshUpdateUsersOfGroup, ins.Data, resource.RedisInstance.RedisDo)
	if err != nil {
		log.SysF("updateUsersOfGroup err: %s", err)
		return err
	}
	return nil
}

func updateUsersOfGroupToDB(dbName, group string, cmdsExecutor func(*redis.RedisCommands) error) error {
	waitForDbConnection.Lock()
	defer waitForDbConnection.Unlock()

	session, err := resource.MongoPool.GetSession()
	defer resource.MongoPool.ReturnSession(session, err)
	if err != nil {
		return err
	}
	err = dispatcher.UpdateUsersOfGroup(session, dbName, group, cmdsExecutor)
	if err != nil {
		return err
	}
	return nil
}

func updateUser(msg *nsq.Message) error {
	type InnerSyncMessage struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}
	var ins InnerSyncMessage
	err := json.Unmarshal(msg.Body, &ins)
	if err != nil {
		log.SysF("update user err: %s", err)
		return err
	}
	switch ins.Type {
	case "add", "update":
		// session := resource.MongoPool.GetSession()
		// defer resource.MongoPool.ReturnSession(session)
		// err = dispatcher.AddUsers(session, conf.GetDbName(), bson.M{"_id": ins.Data}, resource.RedisInstance.RedisDoMulti)
		err = updateUserToDB(conf.Get("dbName").(string), bson.M{"_id": ins.Data}, resource.RedisInstance.RedisDoMulti)
		if err != nil {
			log.SysF("update user err: %s", err)
			return err
		}
	case "remove":
		err = lua.RemoveUsersFromRedis([]string{ins.Data}, resource.RedisInstance.RedisDoMulti)
		if err != nil {
			log.SysF("update user err: %s", err)
			return err
		}
	}
	return nil
}

func updateUserToDB(dbName string, query interface{}, cmdsExecutor func(*redis.RedisCommands) error) error {
	waitForDbConnection.Lock()
	defer waitForDbConnection.Unlock()

	session, err := resource.MongoPool.GetSession()
	defer resource.MongoPool.ReturnSession(session, err)
	if err != nil {
		return err
	}
	err = dispatcher.AddUsers(session, dbName, query, cmdsExecutor)
	if err != nil {
		return err
	}
	return nil
}

func updateGroup(msg *nsq.Message) error {
	type InnerSyncMessage struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}
	var ins InnerSyncMessage
	err := json.Unmarshal(msg.Body, &ins)
	if err != nil {
		log.SysF("update group err: %s", err)
		return err
	}
	switch ins.Type {
	case "add":
		// session := resource.MongoPool.GetSession()
		// defer resource.MongoPool.ReturnSession(session)
		// err = dispatcher.AddNewGroup(session, conf.GetDbName(), ins.Data, resource.RedisInstance.RedisDoMulti, resource.RedisInstance.DoScript)
		err = addNewGroupToDB(conf.Get("dbName").(string), ins.Data, resource.RedisInstance.RedisDoMulti, resource.RedisInstance.DoScript)
		if err != nil {
			log.SysF("updateGroup err: %s", err)
			return err
		}
	case "remove":
		err = lua.RemoveGroup(ins.Data, resource.RedisInstance.DoScript)
		if err != nil {
			return err
		}
		//notify node to remove group
		err = dispatcher.NotifyNodeDataRefresh(dispatcher.DataRefreshRemoveGroup, ins.Data, resource.RedisInstance.RedisDo)
		if err != nil {
			log.SysF("updateGroup err: %s", err)
			return err
		}
	case "update": //暂时忽略,当前用不到具体的支部信息
	}
	return nil
}

func addNewGroupToDB(dbName, groupID string, cmdsExecutor func(*redis.RedisCommands) error, scriptExecutor dispatcher.ScriptExecutor) error {
	waitForDbConnection.Lock()
	defer waitForDbConnection.Unlock()

	session, err := resource.MongoPool.GetSession()
	defer resource.MongoPool.ReturnSession(session, err)
	if err != nil {
		return err
	}
	err = dispatcher.AddNewGroup(session, dbName, groupID, cmdsExecutor, scriptExecutor)
	if err != nil {
		return err
	}
	return nil
}

func startNsqConsumer(nsqlookupdAddress, topic string, handler nsq.Handler) {
	config := nsq.NewConfig()
	config.LookupdPollInterval = lookupdPollInterval
	// var err error

	consumer, err := nsq.NewConsumer(topic, nsqChannel, config)
	if err != nil {
		panic(err)
	}

	consumer.AddHandler(handler)

	err = consumer.ConnectToNSQLookupd(nsqlookupdAddress)
	// err = consumer.ConnectToNSQD("127.0.0.1:4150")
	if err != nil {
		panic("ConnectToNSQLookupd failed: " + err.Error())
	}
}
