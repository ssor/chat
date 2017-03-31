package hub

import "testing"
import "fmt"

func TestRemoveUser(t *testing.T) {
	usrs := UserList{}
	for index := 0; index < 5; index++ {
		usrs = append(usrs, user(&coreUser{fmt.Sprintf("user_%d", index)}))
	}

}

// coreUser
type coreUser struct {
	id string
}

func (cu *coreUser) GetID() string {
	return cu.id
}
func (cu *coreUser) AddMessageToCache(msgID string) {
}
func (cu *coreUser) RemoveRecordCache(id string) {
}
func (cu *coreUser) SendMessage() {
}

func (cu *coreUser) Release() {
}
func (cu *coreUser) RemoveOffline() bool {
	return true
}
func (cu *coreUser) IsOnline() bool {
	return true
}
