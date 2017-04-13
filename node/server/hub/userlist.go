package hub

import (
	user "github.com/ssor/chat/node/server/user"
)

// UserList list of users used in core
type UserList []*user.User

// ToUserList converts interfaces to userList
func ToUserList(users ...interface{}) UserList {
	list := UserList{}
	for _, u := range users {
		list = append(list, (u).(*user.User))
	}
	return list
}

func (ul UserList) add(u *user.User) (list UserList) {
	return append(ul, u)
}

func (ul UserList) find(id string) (u *user.User) {
	for _, _user := range ul {
		if _user.GetID() == id {
			u = _user
			break
		}
	}
	return
}

func (ul UserList) remove(id string) (list UserList, u *user.User) {
	for index, _user := range ul {
		if _user.GetID() == id {
			u = _user
			list = ul.removeByIndex(index)
			break
		}
	}
	return
}

func (ul UserList) removeByIndex(index int) (list UserList) {
	if index >= len(ul) {
		list = ul[:]
	} else {
		list = append(ul[:index], ul[index+1:]...)
	}
	return
}
