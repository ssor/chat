package core

// UserList list of users used in core
type UserList []user

type user interface {
	GetID() string
	AddMessageToCache(message)                // add message to send to client to user's cache
	BroadcastMessage()                        // remind user to start send stored message to client
	SetMessagePopHandler(func(message) error) // if a new message comes in, use this handler
	Release()                                 // release user's resouce if it has
}

// ToUserList converts interfaces to userList
func ToUserList(users ...interface{}) UserList {
	list := UserList{}
	for _, u := range users {
		list = append(list, (u).(user))
	}
	return list
}

func (ul UserList) add(u user) (list UserList) {
	return append(ul, u)
}

func (ul UserList) find(id string) (u user) {
	for _, _user := range ul {
		if _user.GetID() == id {
			u = _user
			break
		}
	}
	return
}

func (ul UserList) remove(id string) (list UserList, u user) {
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
	if (index + 1) >= len(ul) {
		list = ul[:index]
	} else {
		list = append(ul[:index], ul[index+1:]...)
	}
	return
}
