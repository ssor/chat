package detail

type FakeUser struct {
	id string
}

func NewFakeUser(id string) *FakeUser {
	return &FakeUser{id: id}
}

func (fui *FakeUser) GetID() string {
	return fui.id
}

func (fui *FakeUser) GetrName() string {
	return "FakeUser"
}

func (fui *FakeUser) IsFake() bool {
	return true
}
