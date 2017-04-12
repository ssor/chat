package group

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

// Group is a group of fakeusers
type Group struct {
	users UserSecretList
}

func NewGroup(count int, groupID, nodeHost string) *Group {
	group := &Group{
		users: UserSecretList{},
	}
	for index := 0; index < count; index++ {
		group.addUserSecret(FormatFakeUserID(index), groupID, nodeHost)
	}
	return group
}

func (g *Group) addUserSecret(id string, groupID, nodeHost string) {
	us := NewUserSecret(id, groupID, nodeHost)
	g.users = append(g.users, us)
}

func (g *Group) SendMessage(content string) {
	for _, us := range g.users {
		err := us.sendMessage(content)
		if err != nil {
			fmt.Printf("send message to user %s err: %s", us.id, err.Error())
		}
	}
}
func (g *Group) DumpMessage() {
	for _, us := range g.users {
		fmt.Println("user -> ", us.id)
		spew.Dump(us.messages)
		// for _, msg := range us.messages {
		// 	fmt.Println(msg.String())
		// }
	}
}
