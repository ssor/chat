package detail

import db "xsbPro/xsbdb"

type RealUser struct {
	*db.User
}

func NewRealUser(u *db.User) *RealUser {
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
