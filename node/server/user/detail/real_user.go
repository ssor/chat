package detail

import (
	"github.com/ssor/chat/mongo"
)

type RealUser struct {
	*mongo.User
}

func NewRealUser(u *mongo.User) *RealUser {
	return &RealUser{u}
}

func (rui *RealUser) GetID() string {
	return rui.ID
}

func (rui *RealUser) GetName() string {
	return rui.Name
}

func (rui *RealUser) IsFake() bool {
	return false
}
