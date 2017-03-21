package tests

import (
	"testing"
	dispatcher "xsbPro/chatDispatcher/dispatcher"
	"xsbPro/chatDispatcher/lua"
	. "xsbPro/chatDispatcher/resource"
	"xsbPro/common"
	db "xsbPro/xsbdb"

	mgo "gopkg.in/mgo.v2"

	"fmt"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	users           db.UserArray
	groups          []*db.Group
	group_users_map map[string]db.UserArray
	nodes           []*dispatcher.NodeInfo
)

func init() {
	users = db.UserArray{}
	groups = []*db.Group{}
	group_users_map = make(map[string]db.UserArray)
	//prepare some data
	// 5 groups created, and each group has 3 users
	for groupIndex := 1; groupIndex <= 5; groupIndex++ {
		groups = append(groups, &db.Group{
			ID:   fmt.Sprintf("group_id_%d", groupIndex),
			Name: fmt.Sprintf("group_name_%d", groupIndex),
		})
	}

	for _, group := range groups {
		group_users_map[group.ID] = db.UserArray{}
		for userIndex := 1; userIndex <= 3; userIndex++ {
			user := &db.User{
				ID:    fmt.Sprintf("user_id_%d_%s", userIndex, group.ID),
				Name:  fmt.Sprintf("user_name_%d_%s", userIndex, group.Name),
				Group: group.ID,
			}
			users = append(users, user)
			group_users_map[group.ID] = append(group_users_map[group.ID], user)
		}
	}

	fillBasicDataToMongo(groups, users)
	// session := Mongo_pool.GetSession()
	// defer Mongo_pool.ReturnSession(session)

	// err := clearDB(session, conf.GetDbName())
	// Panic_error(err)
	// err = insertGroupsToDB(session, conf.GetDbName(), groups)
	// Panic_error(err)
	// err = insertUsersToDB(session, conf.GetDbName(), users)
	// Panic_error(err)

	nodes = []*dispatcher.NodeInfo{}
	nodes = append(nodes, dispatcher.NewNodeInfo("lan_01", "wan_01", 4))
	nodes = append(nodes, dispatcher.NewNodeInfo("lan_02", "wan_02", 4))
	nodes = append(nodes, dispatcher.NewNodeInfo("lan_03", "wan_03", 4))
}

func fillBasicDataToMongo(groups []*db.Group, users db.UserArray) {

	session, err := Mongo_pool.GetSession()
	if err != nil {
		panic(err)
	}
	defer Mongo_pool.ReturnSession(session, err)

	err = clearDB(session, conf.GetDbName())
	Panic_error(err)
	err = insertGroupsToDB(session, conf.GetDbName(), groups)
	Panic_error(err)
	err = insertUsersToDB(session, conf.GetDbName(), users)
	Panic_error(err)
}

func TestLoadDataFromDB(t *testing.T) {
	Convey("fill data to redis from mongo", t, func() {
		session, err := Mongo_pool.GetSession()
		if err != nil {
			panic(err)
		}
		defer Mongo_pool.ReturnSession(session, err)

		Convey("fill data to redis", func() {
			err := dispatcher.FillDataToRedisFromMongo(session, conf.GetDbName(), Redis_instance.RedisDoMulti, Redis_instance.DoScript)
			So(err, ShouldEqual, nil)
		})

		Convey("get data from redis", func() {
			for group_id, users_in_group := range group_users_map {
				users_from_cache, err := lua.GetGroupUsersFromCache(group_id, Redis_instance.DoScript)
				Panic_error(err)
				// log.Println("src:")
				// for _, user := range users_in_group {
				// 	log.Println("user id -> ", user.ID)
				// }
				// log.Println("---------------------")
				// log.Println("cached:")
				// for _, user := range users_from_cache {
				// 	log.Println("user id -> ", user.ID)
				// }
				So(usersEqual(users_in_group, users_from_cache), ShouldEqual, true)
			}
		})
	})
}

func TestNodeRegister(t *testing.T) {
	Convey("add node", t, func() {
		for index := range nodes {
			err := dispatcher.RegisterToNodeCenter(nodes[index], Redis_instance.RedisDo)
			Panic_error(err)
		}

		for index := range nodes {
			ni_cached, err := dispatcher.GetNodeInfoByKey(nodes[index].Key, Redis_instance.DoScript)
			Panic_error(err)
			So(ni_cached.Equal(nodes[index]), ShouldEqual, true)
		}

	})
	// Convey("node update", t, func() {

	// })
	node0 := nodes[0]
	node1 := nodes[1]
	node2 := nodes[2]
	Convey("node update capability, and distribute groups to node", t, func() {
		err := dispatcher.UpdateNodeCapacity(node0.LanHost, node0.Max, Redis_instance.DoScript)
		count, err := dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, len(groups)-node0.Max)

		db_groups, err := lua.GetGroupsOnNode(node0.LanHost, Redis_instance.DoScript)
		Panic_error(err)
		So(len(db_groups), ShouldEqual, node0.Max)

		db_users, err := lua.GetGroupUsersOnNode(db_groups[0].ID, node0.LanHost, Redis_instance.DoScript)
		Panic_error(err)
		So(usersEqual(db_users, group_users_map[db_groups[0].ID]), ShouldEqual, true)

		db_users_null, err := lua.GetGroupUsersOnNode(db_groups[0].ID, node1.LanHost, Redis_instance.DoScript)
		Panic_error(err)
		So(db_users_null, ShouldEqual, nil)
		// group_id_list, err := lua.GetDispatchedGroupsOfNode(node0.LanHost, Redis_instance.DoScript)
		// Panic_error(err)
		// So(len(group_id_list), ShouldEqual, node0.Max)

		err = dispatcher.RemoveNode(node0.LanHost, Redis_instance.DoScript)
		Panic_error(err)
		count, err = dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, len(groups))

		err = dispatcher.RegisterToNodeCenter(node0, Redis_instance.RedisDo)
		Panic_error(err)
		err = dispatcher.UpdateNodeCapacity(node0.LanHost, node0.Max, Redis_instance.DoScript)
		count, err = dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, len(groups)-node0.Max)

		err = dispatcher.UpdateNodeCapacity(node1.LanHost, node1.Max, Redis_instance.DoScript)
		count, err = dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, 0)

		err = dispatcher.RemoveNode(node1.LanHost, Redis_instance.DoScript)
		Panic_error(err)
		count, err = dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, len(groups)-node0.Max)

		err = dispatcher.RegisterToNodeCenter(node1, Redis_instance.RedisDo)
		Panic_error(err)
		err = dispatcher.UpdateNodeCapacity(node1.LanHost, node1.Max, Redis_instance.DoScript)
		count, err = dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, 0)

		err = dispatcher.RegisterToNodeCenter(node2, Redis_instance.RedisDo)
		Panic_error(err)
		err = dispatcher.UpdateNodeCapacity(node2.LanHost, node2.Max, Redis_instance.DoScript)
		Panic_error(err)

		err = dispatcher.RemoveNode(node1.LanHost, Redis_instance.DoScript)
		Panic_error(err)
		count, err = dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, 0)
	})

	new_groups := []*db.Group{}
	Convey("new group added", t, func() {
		for index := 0; index < 4; index++ {
			new_groups = append(new_groups, &db.Group{ID: fmt.Sprintf("new_group_%d", index), Name: fmt.Sprintf("new_group_name_%d", index)})
		}
		for _, group := range new_groups {
			err := dispatcher.FillNewGroupToRedis(group, Redis_instance.DoScript)
			Panic_error(err)
		}

		count, err := dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, 1)

	})

	Convey("remove group", t, func() {
		err := lua.RemoveGroupFromRedis(new_groups[0].ID, Redis_instance.DoScript)
		Panic_error(err)

		count, err := dispatcher.GetUnloadGroupCount(Redis_instance.DoScript)
		Panic_error(err)
		So(count, ShouldEqual, 0)
	})

	Convey("handle login route request", t, func() {
		for _, group := range groups {
			wan, err := dispatcher.GetNodeByGroup(group.ID, Redis_instance.DoScript)
			Panic_error(err)
			So(wan, ShouldContainSubstring, "wan_")
		}
	})

	Convey("reload data from db", t, func() {
		users = db.UserArray{}
		groups = []*db.Group{}
		group_users_map = make(map[string]db.UserArray)
		//prepare some data
		// 5 groups created, and each group has 3 users
		for groupIndex := range []int{1, 2, 3, 4, 6} {
			groups = append(groups, &db.Group{
				ID:   fmt.Sprintf("group_id_%d", groupIndex),
				Name: fmt.Sprintf("group_name_%d", groupIndex),
			})
		}

		for _, group := range groups {
			group_users_map[group.ID] = db.UserArray{}
			for userIndex := range []int{4, 5, 6} {
				user := &db.User{
					ID:    fmt.Sprintf("user_id_%d_%s", userIndex, group.ID),
					Name:  fmt.Sprintf("user_name_%d_%s", userIndex, group.Name),
					Group: group.ID,
				}
				users = append(users, user)
				group_users_map[group.ID] = append(group_users_map[group.ID], user)
			}
		}
		fillBasicDataToMongo(groups, users)
		session, err := Mongo_pool.GetSession()
		defer Mongo_pool.ReturnSession(session, err)

		Convey("fill data to redis", func() {
			err := dispatcher.FillDataToRedisFromMongo(session, conf.GetDbName(), Redis_instance.RedisDoMulti, Redis_instance.DoScript)
			So(err, ShouldEqual, nil)
		})

		Convey("get data from redis", func() {
			for group_id, users_in_group := range group_users_map {
				users_from_cache, err := lua.GetGroupUsersFromCache(group_id, Redis_instance.DoScript)
				Panic_error(err)
				// log.Println("src:")
				// for _, user := range users_in_group {
				// 	log.Println("user id -> ", user.ID)
				// }
				// log.Println("---------------------")
				// log.Println("cached:")
				// for _, user := range users_from_cache {
				// 	log.Println("user id -> ", user.ID)
				// }
				So(usersEqual(users_in_group, users_from_cache), ShouldEqual, true)
			}
		})
		wan, err := dispatcher.GetNodeByGroup(groups[len(groups)-1].ID, Redis_instance.DoScript)
		Panic_error(err)
		So(wan, ShouldContainSubstring, "wan_")
	})

	// Convey("notify node that group's user info changed", t, func() {

	// })
}

func TestCheckNodeStatus(t *testing.T) {
	Convey("check node status", t, func() {

	})
}

func usersEqual(src, dest db.UserArray) bool {
	if len(src) != len(dest) {
		return false
	}

	has_user := func(users db.UserArray, user *db.User) bool {
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
		if has_user(dest, user) == false {
			return false
		}
	}
	return true
}

func clearDB(session *mgo.Session, db_name string) error {
	err := session.DB(db_name).DropDatabase()
	return err
}

func insertDataToDB(session *mgo.Session, db_name, collection string, objs ...interface{}) error {
	for _, obj := range objs {
		err := session.DB(db_name).C(collection).Insert(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertUsersToDB(session *mgo.Session, db_name string, users db.UserArray) error {
	for _, user := range users {
		err := insertDataToDB(session, db_name, common.Collection_userinfo, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertGroupsToDB(session *mgo.Session, db_name string, groups []*db.Group) error {
	for _, group := range groups {
		if err := insertDataToDB(session, db_name, common.Collection_group, group); err != nil {
			return err
		}
	}
	return nil
}

// func GetUsersFromDB(session *mgo.Session, db_name string, query interface{}) (db.UserArray, error) {

// 	if session == nil {
// 		return nil, fmt.Errorf("db session should not be nil")
// 	}

// 	var users_array db.UserArray
// 	err := session.DB(db_name).C(common.Collection_userinfo).Find(query).All(&users_array)
// 	if err != nil {
// 		log.SysF("getUsersFromDB error: %s", err)
// 		return nil, err
// 	}

// 	return users_array, nil
// }

// func GetGroupsFromDB(session *mgo.Session, db_name string, query interface{}) ([]*db.Group, error) {

// 	if session == nil {
// 		return nil, fmt.Errorf("db session should not be nil")
// 	}

// 	var groups []*db.Group
// 	err := session.DB(db_name).C(common.Collection_group).Find(query).All(&groups)
// 	if err != nil {
// 		log.SysF("GetGroupsFromDB error: %s", err)
// 		return nil, err
// 	}

// 	return groups, nil
// }
