package core

import (
	"runtime"
	"sync"
	"time"
	"xsbPro/log"
)

//HubManager hold a list of hubs
// It also has a ticker to remind hubs to send msg intervally
type HubManager struct {
	Hubs    HubList
	hubSync sync.Mutex
}

// NewHubManager init a empty manager, but need run
func NewHubManager() *HubManager {
	hm := &HubManager{
		Hubs:    HubList{},
		hubSync: sync.Mutex{},
	}
	go hm.Run()
	return hm
}

// AddHub add a new hub with id and users in group
func (hm *HubManager) AddHub(group string, users UserList) *Hub {
	hm.hubSync.Lock()
	defer hm.hubSync.Unlock()

	if hm.Hubs.Contains(group) == true {
		return nil
	}
	hub := newHub(group, users)
	hm.addHub(hub)
	return hub
}

func (hm *HubManager) addHub(hub *Hub) {
	hm.Hubs = hm.Hubs.add(hub)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackInfo := string(stack())
				log.MustF("hub stop running: %s", stackInfo)
				log.LogToFile(stackInfo)
			}
			hm.addHub(hub)
		}()

		hub.run()
	}()
}

// func (hm *HubManager) GetHubs() map[string]*Hub {
// 	return hm.Hubs.Items()
// }

// RemoveHub removes hub with group id
func (hm *HubManager) RemoveHub(id string) error {
	hm.hubSync.Lock()
	defer hm.hubSync.Unlock()

	list, hub := hm.Hubs.remove(id)
	if hub == nil {
		return nil
	}

	hub.close()
	hm.Hubs = list
	return nil
}

// Run starts loop
func (hm *HubManager) Run() {
	messageTicker := time.NewTicker(2 * time.Second) //出发消息发送轮询事件
	for {
		select {
		// case <-hm.chCommand:
		case <-messageTicker.C:
			for _, h := range hm.Hubs {
				h.eventSendMessage <- 0
			}
		}
	}
}

func stack() []byte {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return buf[:n]
}
