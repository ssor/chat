package server

import (
	"fmt"

	"github.com/ssor/chat/lua"
	"github.com/ssor/chat/mongo"
	"github.com/ssor/chat/node/server/hub"
	"github.com/ssor/chat/node/server/user"
	"github.com/ssor/chat/node/server/user/detail"
	"github.com/ssor/log"
)

//将 redis 中的数据和本地数据进行同步,  删除已经不存在的, 更新已有的 Hub 中的用户信息
func refreshGroupsFromRedis(node string, scriptExecutor lua.ScriptExecutor) error {
	groups, err := lua.GetGroupsOnNode(node, scriptExecutor)
	if err != nil {
		return err
	}
	log.InfoF("%d groups on node %s", len(groups), node)

	err = removeNotExistingHubs(groups)
	if err != nil {
		return err
	}
	err = refreshUsersOfExistingHubs(groups, scriptExecutor)
	if err != nil {
		return err
	}

	return nil
}

func removeNotExistingHubs(groups []*mongo.Group) error {
	hubsCurrent := serverInstance.hubManager.Hubs
	for _, h := range hubsCurrent {
		if ifGroupExists(groups, h.GetID()) == false {
			err := serverInstance.hubManager.RemoveHub(h.GetID())
			if err != nil {
				return err
			}
			log.InfoF("group %s has been removed from db")
		}
	}
	return nil
}

func refreshUsersOfExistingHubs(groups []*mongo.Group, scriptExecutor lua.ScriptExecutor) error {
	hubsCurrent := serverInstance.hubManager.Hubs
	for _, h := range hubsCurrent {
		if ifGroupExists(groups, h.GetID()) {
			err := refreshHubUsers(h.GetID(), scriptExecutor)
			if err != nil {
				return err
			}
			log.InfoF("users of group %s has been updated")
		}
	}
	return nil
}

func ifGroupExists(groups []*mongo.Group, id string) bool {
	for _, group := range groups {
		if group.ID == id {
			return true
		}
	}
	return false
}

// loadGroupFromRedis load group users into memory
func loadGroupFromRedis(group, node string, scriptExecutor lua.ScriptExecutor) (*hub.Hub, error) {
	usersArray, err := lua.GetGroupUsersOnNode(group, node, scriptExecutor)
	if err != nil {
		return nil, err
	}
	newHub := serverInstance.hubManager.AddHub(group, nil)
	// users := hub.ToUserList(usersArray)
	newHub.AddUser(convertDbUserToHubUser(usersArray, newHub)...)
	return newHub, nil
}

// func refreshUsers(id string, scriptExecutor lua.ScriptExecutor) error {
// 	hub := hm.Hubs.Get(id)
// 	if hub == nil {
// 		return nil
// 	}
// 	return hub.RefreshUsers(scriptExecutor)
// }

// RefreshHubUsers refresh users in hub
func refreshHubUsers(hubID string, scriptExecutor lua.ScriptExecutor) error {
	h := serverInstance.findHub(hubID)
	if h == nil {
		return fmt.Errorf("hub %s not found, so cannot refresh users in it", hubID)
	}

	users, err := lua.GetGroupUsers(hubID, scriptExecutor)
	if err != nil {
		return err
	}
	// addNewUserToHub(users, h)
	usersNotInHub := filterUsersNotInHub(users, h)
	addNewUserToHub(h, convertDbUserToHubUser(usersNotInHub, h)...)
	removeUsersNotInHub(users, h)
	return nil
}

func removeUsersNotInHub(usersShouldIn mongo.UserArray, h *hub.Hub) {
	//some user removed from this group
	findUser := func(users mongo.UserArray, id string) *mongo.User {
		for _, user := range users {
			if user.ID == id {
				return user
			}
		}
		return nil
	}
	for _, user := range h.GroupUsers {
		if findUser(usersShouldIn, user.GetID()) == nil {
			h.RemoveUser(user)
		}
	}

}

func filterUsersNotInHub(users mongo.UserArray, h *hub.Hub) mongo.UserArray {
	list := mongo.UserArray{}

	for _, dbUser := range users {
		if h.FindUser(dbUser.ID) == nil {
			list = append(list, dbUser)
		}
	}
	return list
}

func addNewUserToHub(h *hub.Hub, users ...*user.User) {
	h.AddUser(users...)
	//new user added to this group
	// for _, dbUser := range users {
	// 	if h.FindUser(dbUser.ID) == nil {
	// 		ru := detail.NewRealUser(dbUser)
	// 		h.AddUser(user.NewUser(ru, h))
	// 	}
	// }
}

func convertDbUserToHubUser(users mongo.UserArray, h *hub.Hub) []*user.User {
	list := []*user.User{}
	if users == nil {
		return list
	}

	for _, dbUser := range users {
		ru := detail.NewRealUser(dbUser)
		list = append(list, user.NewUser(ru, h))
	}
	return list
}

// func (h *Hub) acceptInterview(questionnaire *Questionnaire) {
// 	h.investigate <- questionnaire
// }

// RefreshAll refresh all groups on this node
func RefreshAll(scriptExecutor lua.ScriptExecutor) error {
	err := refreshGroupsFromRedis(nodeID, scriptExecutor)
	if err != nil {
		return err
	}
	return nil
}

// RemoveGroup removes specified group
func RemoveGroup(group string) error {
	return serverInstance.hubManager.RemoveHub(group)
}

// RefreshGroupUsers refresh users in hub
func RefreshGroupUsers(group string, scriptExecutor lua.ScriptExecutor) error {
	return refreshHubUsers(group, scriptExecutor)
}
