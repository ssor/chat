package mongo

import "fmt"

type Group struct {
	ID         string `bson:"_id"`
	Name       string `bson:"name"`
	Company    string `bson:"company"`
	CreateTime string `bson:"createtime"`
}

func (this *Group) String() string {
	return fmt.Sprintf("ID: %s Name: %s Company: %s", this.ID, this.Name, this.Company)
}

type DBGroupArray []*Group
