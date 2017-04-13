package tests

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/ssor/chat/dispatcher/dispatcher"
	"github.com/ssor/chat/dispatcher/resource"
	"github.com/ssor/chat/lua"
	"github.com/ssor/chat/mongo"
	mgo "gopkg.in/mgo.v2"
)

var (
	users         mongo.UserArray
	groups        []*mongo.Group
	groupUsersMap map[string]mongo.UserArray
	nodes         []*dispatcher.NodeInfo
)

func init() {
	users = mongo.UserArray{}
	groups = []*mongo.Group{}
	groupUsersMap = make(map[string]mongo.UserArray)
	//prepare some data
	// 5 groups created, and each group has 3 users
	for groupIndex := 1; groupIndex <= 5; groupIndex++ {
		groups = append(groups, &mongo.Group{
			ID:   fmt.Sprintf("group_id_%d", groupIndex),
			Name: fmt.Sprintf("group_name_%d", groupIndex),
		})
	}

	for _, group := range groups {
		groupUsersMap[group.ID] = mongo.UserArray{}
		for userIndex := 1; userIndex <= 3; userIndex++ {
			user := &mongo.User{
				ID:    fmt.Sprintf("user_id_%d_%s", userIndex, group.ID),
				Name:  fmt.Sprintf("user_name_%d_%s", userIndex, group.Name),
				Group: group.ID,
			}
			users = append(users, user)
			groupUsersMap[group.ID] = append(groupUsersMap[group.ID], user)
		}
	}

	fillBasicDataToMongo(groups, users)
	// session := resource.MongoPool.GetSession()
	// defer resource.MongoPool.ReturnSession(session)

	// err := clearDB(session, conf.Get("dbName").(string))
	// panicError(err)
	// err = insertGroupsToDB(session, conf.Get("dbName").(string), groups)
	// panicError(err)
	// err = insertUsersToDB(session, conf.Get("dbName").(string), users)
	// panicError(err)

	nodes = []*dispatcher.NodeInfo{}
	nodes = append(nodes, dispatcher.NewNodeInfo("lan_01", "wan_01", 4))
	nodes = append(nodes, dispatcher.NewNodeInfo("lan_02", "wan_02", 4))
	nodes = append(nodes, dispatcher.NewNodeInfo("lan_03", "wan_03", 4))
}

func fillBasicDataToMongo(groups []*mongo.Group, users mongo.UserArray) {

	session, err := resource.MongoPool.GetSession()
	if err != nil {
		panic(err)
	}
	defer resource.MongoPool.ReturnSession(session, err)

	err = clearDB(session, conf.Get("dbName").(string))
	panicError(err)
	err = insertGroupsToDB(session, conf.Get("dbName").(string), groups)
	panicError(err)
	err = insertUsersToDB(session, conf.Get("dbName").(string), users)
	panicError(err)
}

func TestLoadDataFromDB(t *testing.T) {
	Convey("fill data to redis from mongo", t, func() {
		session, err := resource.MongoPool.GetSession()
		if err != nil {
			panic(err)
		}
		defer resource.MongoPool.ReturnSession(session, err)

		Convey("fill data to redis", func() {
			err := dispatcher.FillDataToRedisFromMongo(session, conf.Get("dbName").(string), resource.RedisInstance.DoMulti, resource.RedisInstance.DoScript)
			So(err, ShouldEqual, nil)
		})

		Convey("get data from redis", func() {
			for groupID, usersInGroup := range groupUsersMap {
				usersFromCache, err := lua.GetGroupUsers(groupID, resource.RedisInstance.DoScript)
				panicError(err)
				// log.Println("src:")
				// for _, user := range usersInGroup {
				// 	log.Println("user id -> ", user.ID)
				// }
				// log.Println("---------------------")
				// log.Println("cached:")
				// for _, user := range usersFromCache {
				// 	log.Println("user id -> ", user.ID)
				// }
				So(usersEqual(usersInGroup, usersFromCache), ShouldEqual, true)
			}
		})
	})
}

func TestNodeRegister(t *testing.T) {
	Convey("add node", t, func() {
		for index := range nodes {
			err := dispatcher.RegisterToNodeCenter(nodes[index], resource.RedisInstance.Do)
			panicError(err)
		}

		for index := range nodes {
			niCached, err := dispatcher.GetNodeInfoByKey(nodes[index].Key, resource.RedisInstance.DoScript)
			panicError(err)
			So(niCached.Equal(nodes[index]), ShouldEqual, true)
		}
	})

	node0 := nodes[0]
	node1 := nodes[1]
	node2 := nodes[2]
	Convey("node update capability, and distribute groups to node", t, func() {
		err := dispatcher.UpdateNodeCapacity(node0.LanHost, node0.Max, resource.RedisInstance.DoScript)
		count, err := dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, len(groups)-node0.Max)

		dbGroups, err := lua.GetGroupsOnNode(node0.LanHost, resource.RedisInstance.DoScript)
		panicError(err)
		So(len(dbGroups), ShouldEqual, node0.Max)

		dbUsers, err := lua.GetGroupUsersOnNode(dbGroups[0].ID, node0.LanHost, resource.RedisInstance.DoScript)
		panicError(err)
		So(usersEqual(dbUsers, groupUsersMap[dbGroups[0].ID]), ShouldEqual, true)

		dbUsersNull, err := lua.GetGroupUsersOnNode(dbGroups[0].ID, node1.LanHost, resource.RedisInstance.DoScript)
		panicError(err)
		So(dbUsersNull, ShouldEqual, nil)

		err = dispatcher.RemoveNode(node0.LanHost, resource.RedisInstance.DoScript)
		panicError(err)
		count, err = dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, len(groups))

		err = dispatcher.RegisterToNodeCenter(node0, resource.RedisInstance.Do)
		panicError(err)
		err = dispatcher.UpdateNodeCapacity(node0.LanHost, node0.Max, resource.RedisInstance.DoScript)
		count, err = dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, len(groups)-node0.Max)

		err = dispatcher.UpdateNodeCapacity(node1.LanHost, node1.Max, resource.RedisInstance.DoScript)
		count, err = dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, 0)

		err = dispatcher.RemoveNode(node1.LanHost, resource.RedisInstance.DoScript)
		panicError(err)
		count, err = dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, len(groups)-node0.Max)

		err = dispatcher.RegisterToNodeCenter(node1, resource.RedisInstance.Do)
		panicError(err)
		err = dispatcher.UpdateNodeCapacity(node1.LanHost, node1.Max, resource.RedisInstance.DoScript)
		count, err = dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, 0)

		err = dispatcher.RegisterToNodeCenter(node2, resource.RedisInstance.Do)
		panicError(err)
		err = dispatcher.UpdateNodeCapacity(node2.LanHost, node2.Max, resource.RedisInstance.DoScript)
		panicError(err)

		err = dispatcher.RemoveNode(node1.LanHost, resource.RedisInstance.DoScript)
		panicError(err)
		count, err = dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, 0)
	})

	newGroups := []*mongo.Group{}
	Convey("new group added", t, func() {
		for index := 0; index < 4; index++ {
			newGroups = append(newGroups, &mongo.Group{ID: fmt.Sprintf("new_group_%d", index), Name: fmt.Sprintf("new_group_name_%d", index)})
		}
		for _, group := range newGroups {
			err := lua.FillNewGroupToRedis(group, resource.RedisInstance.DoScript)
			panicError(err)
		}

		count, err := dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, 1)

	})

	Convey("remove group", t, func() {
		err := lua.RemoveGroup(newGroups[0].ID, resource.RedisInstance.DoScript)
		panicError(err)

		count, err := dispatcher.GetUnloadGroupCount(resource.RedisInstance.DoScript)
		panicError(err)
		So(count, ShouldEqual, 0)
	})

	Convey("handle login route request", t, func() {
		for _, group := range groups {
			wan, err := lua.GetNodeByGroup(group.ID, resource.RedisInstance.DoScript)
			panicError(err)
			So(wan, ShouldContainSubstring, "wan_")
		}
	})

	Convey("reload data from db", t, func() {
		users = mongo.UserArray{}
		groups = []*mongo.Group{}
		groupUsersMap = make(map[string]mongo.UserArray)
		//prepare some data
		// 5 groups created, and each group has 3 users
		for groupIndex := range []int{1, 2, 3, 4, 6} {
			groups = append(groups, &mongo.Group{
				ID:   fmt.Sprintf("group_id_%d", groupIndex),
				Name: fmt.Sprintf("group_name_%d", groupIndex),
			})
		}

		for _, group := range groups {
			groupUsersMap[group.ID] = mongo.UserArray{}
			for userIndex := range []int{4, 5, 6} {
				user := &mongo.User{
					ID:    fmt.Sprintf("user_id_%d_%s", userIndex, group.ID),
					Name:  fmt.Sprintf("user_name_%d_%s", userIndex, group.Name),
					Group: group.ID,
				}
				users = append(users, user)
				groupUsersMap[group.ID] = append(groupUsersMap[group.ID], user)
			}
		}
		fillBasicDataToMongo(groups, users)
		session, err := resource.MongoPool.GetSession()
		defer resource.MongoPool.ReturnSession(session, err)

		Convey("fill data to redis", func() {
			err := dispatcher.FillDataToRedisFromMongo(session, conf.Get("dbName").(string), resource.RedisInstance.DoMulti, resource.RedisInstance.DoScript)
			So(err, ShouldEqual, nil)
		})

		Convey("get data from redis", func() {
			for groupID, usersInGroup := range groupUsersMap {
				usersFromCache, err := lua.GetGroupUsers(groupID, resource.RedisInstance.DoScript)
				panicError(err)
				// log.Println("src:")
				// for _, user := range usersInGroup {
				// 	log.Println("user id -> ", user.ID)
				// }
				// log.Println("---------------------")
				// log.Println("cached:")
				// for _, user := range usersFromCache {
				// 	log.Println("user id -> ", user.ID)
				// }
				So(usersEqual(usersInGroup, usersFromCache), ShouldEqual, true)
			}
		})
		wan, err := lua.GetNodeByGroup(groups[len(groups)-1].ID, resource.RedisInstance.DoScript)
		panicError(err)
		So(wan, ShouldContainSubstring, "wan_")
	})

	// Convey("notify node that group's user info changed", t, func() {

	// })
}

func TestCheckNodeStatus(t *testing.T) {
	Convey("check node status", t, func() {

	})
}

func usersEqual(src, dest mongo.UserArray) bool {
	if len(src) != len(dest) {
		return false
	}

	hasUser := func(users mongo.UserArray, user *mongo.User) bool {
		for _, u := range users {
			if u.ID == user.ID && u.Chief == user.Chief && u.Index == user.Index &&
				u.UserType == user.UserType && u.Gender == user.Gender &&
				u.Actived == user.Actived {
				return true
			}
		}
		return false
	}

	for _, user := range src {
		if hasUser(dest, user) == false {
			return false
		}
	}
	return true
}

func clearDB(session *mgo.Session, dbName string) error {
	err := session.DB(dbName).DropDatabase()
	return err
}

func insertDataToDB(session *mgo.Session, dbName, collection string, objs ...interface{}) error {
	for _, obj := range objs {
		err := session.DB(dbName).C(collection).Insert(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertUsersToDB(session *mgo.Session, dbName string, users mongo.UserArray) error {
	for _, user := range users {
		err := insertDataToDB(session, dbName, mongo.CollectionUserinfo, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertGroupsToDB(session *mgo.Session, dbName string, groups []*mongo.Group) error {
	for _, group := range groups {
		if err := insertDataToDB(session, dbName, mongo.CollectionGroup, group); err != nil {
			return err
		}
	}
	return nil
}

// func GetUsersFromDB(session *mgo.Session, dbName string, query interface{}) (db.UserArray, error) {

// 	if session == nil {
// 		return nil, fmt.Errorf("db session should not be nil")
// 	}

// 	var users_array db.UserArray
// 	err := session.DB(dbName).C(common.Collection_userinfo).Find(query).All(&users_array)
// 	if err != nil {
// 		log.SysF("getUsersFromDB error: %s", err)
// 		return nil, err
// 	}

// 	return users_array, nil
// }

// func GetGroupsFromDB(session *mgo.Session, dbName string, query interface{}) ([]*db.Group, error) {

// 	if session == nil {
// 		return nil, fmt.Errorf("db session should not be nil")
// 	}

// 	var groups []*db.Group
// 	err := session.DB(dbName).C(common.Collection_group).Find(query).All(&groups)
// 	if err != nil {
// 		log.SysF("GetGroupsFromDB error: %s", err)
// 		return nil, err
// 	}

// 	return groups, nil
// }
