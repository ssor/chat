package user

type fakeUserInfo struct {
	id string
}

func newFakeUserInfo(id string) *fakeUserInfo {
	return &fakeUserInfo{id: id}
}

func (fui *fakeUserInfo) GetUserID() string {
	return fui.id
}

func (fui *fakeUserInfo) GetUserName() string {
	return "FakeUser"
}

func (fui *fakeUserInfo) IsFake() bool {
	return true
}
