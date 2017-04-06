package hub

import (
	"runtime"
	"sync"
	"time"
	"xsbPro/log"
)

//Manager hold a list of hubs
// It also has a ticker to remind hubs to send msg intervally
type Manager struct {
	Hubs    List
	hubSync sync.Mutex
}

// NewManager init a empty manager, but need run
func NewManager() *Manager {
	hm := &Manager{
		Hubs:    List{},
		hubSync: sync.Mutex{},
	}
	go hm.Run()
	return hm
}

// AddHub add a new hub with id and users in group
func (hm *Manager) AddHub(group string, users UserList) *Hub {
	hm.hubSync.Lock()
	defer hm.hubSync.Unlock()

	if users == nil {
		users = UserList{}
	}

	hub := hm.Hubs.Find(group)
	if hub == nil {
		hub = newHub(group, users)
		hm.addHub(hub)
	} else {
		for _, u := range users {
			hub.AddUser(u)
		}
	}
	return hub
}

func (hm *Manager) addHub(hub *Hub) {
	hm.Hubs = hm.Hubs.add(hub)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackInfo := string(stack())
				log.MustF("hub stop running: %s", stackInfo)
				log.ToFile(stackInfo)
			}
			hm.addHub(hub)
		}()

		hub.run()
	}()
}

// RemoveHub removes hub with group id
func (hm *Manager) RemoveHub(id string) error {
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
func (hm *Manager) Run() {
	messageTicker := time.NewTicker(3 * time.Second) //出发消息发送轮询事件
	for {
		select {
		// case <-hm.chCommand:
		case <-messageTicker.C:
			for _, h := range hm.Hubs {
				// h.eventSendMessage <- 0
				h.sendMessge()
			}
		}
	}
}

func stack() []byte {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return buf[:n]
}
