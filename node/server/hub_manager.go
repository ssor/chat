package server

import (
	"runtime"
	"time"
	"xsbPro/chatDispatcher/lua"
	"xsbPro/log"
)

//HubManager use
type HubManager struct {
	Hubs      *SafeHubList
	chCommand chan int //接收一些特殊命令
}

func (hm *HubManager) Add(hub *Hub) {
	hm.Hubs.Set(hub.group, hub)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackInfo := string(stack())
				log.MustF("hub stop running: %s", stackInfo)
				log.LogToFile(stackInfo)
			}
			hm.Add(hub)
		}()

		hub.run()
	}()
	// go hub.run()
}

func (hm *HubManager) GetHubs() map[string]*Hub {
	return hm.Hubs.Items()
}

//将 redis 中的数据和本地数据进行同步
func (hm *HubManager) refreshGroupsFromRedis(node string, scriptExecutor lua.ScriptExecutor) error {
	groups, err := lua.GetGroupsOnNode(node, scriptExecutor)
	if err != nil {
		return err
	}
	log.InfoF("%d groups on node %s", len(groups), node)
	//将不应存在的 hub 移除
	//更新现存的 hub 中的用户数据
	func_exists_in_groups := func(group_id string) bool {
		for _, group := range groups {
			if group.ID == group_id {
				return true
			}
		}
		return false
	}
	hubs_current := hm.GetHubs()
	for id := range hubs_current {
		if func_exists_in_groups(id) == false {
			err = hm.RemoveHub(id)
			if err != nil {
				return err
			}
			log.InfoF("group %s has been removed from db")
		} else {
			err = hm.RefreshUsers(id, scriptExecutor)
			if err != nil {
				return err
			}
			log.InfoF("users of group %s has been updated")
		}
	}

	return nil
}

func (hm *HubManager) loadGroupFromRedis(group, node string, scriptExecutor lua.ScriptExecutor) (*Hub, error) {
	users := NewSafeUserList()

	users_array, err := lua.GetGroupUsersOnNode(group, node, scriptExecutor)
	if err != nil {
		return nil, err
	}

	for _, user := range users_array {
		users.Set(user.ID, NewUser(NewRealUserInfo(user), nil))
	}
	nh := NewHub(group, users)
	hm.Add(nh)
	return nh, nil
}
func (hm *HubManager) RefreshUsers(id string, scriptExecutor lua.ScriptExecutor) error {
	hub := hm.Hubs.Get(id)
	if hub == nil {
		return nil
	}
	return hub.RefreshUsers(scriptExecutor)
}

func (hm *HubManager) RemoveHub(id string) error {
	hub := hm.Hubs.Get(id)
	if hub == nil {
		return nil
	}

	hub.close()
	hm.Hubs.Delete(id)
	return nil
}

func (hm *HubManager) Run() {

	messageTicker := time.NewTicker(2 * time.Second) //出发消息发送轮询事件

	for {
		select {
		case <-hm.chCommand:
		case <-messageTicker.C:

			for _, h := range hm.Hubs.Items() {
				h.eventSendMessage <- 0
			}
		}
	}
}

func NewHubManager() *HubManager {
	return &HubManager{
		Hubs: NewSafeHubList(),
	}
}

func stack() []byte {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return buf[:n]
}
