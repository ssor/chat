package server

import (
	"github.com/ssor/chat/node/server/hub"
)

type server struct {
	hubManager *hub.Manager
}

func newServer() *server {
	s := &server{
		hubManager: hub.NewManager(),
	}
	go s.hubManager.Run()
	return s
}

func (s *server) findHub(id string) *hub.Hub {
	return s.hubManager.Hubs.Find(id)
}
