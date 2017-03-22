package scripts

import (
	"fmt"
)

const (
	keyFormatGroupUserRelation = "group->users:%s"
	keyFormatGroup             = "group:%s"
	keyFormatUser              = "user:%s"
	// ip:{ip} max:{max} current:{current},记录节点本身的信息
	hashNode = "node->%s"
)

func FormatNodeInfoKey(lan string) string {
	return fmt.Sprintf(hashNode, lan)
}

// func Format_group_node_map_key(group, lan, wan string) string {
// 	return fmt.Sprintf(common.String_group_node_map, group, lan, wan)
// }

func ForamtGroupKey(group string) string {
	return fmt.Sprintf(keyFormatGroup, group)
}

func FormatUserKey(user string) string {
	return fmt.Sprintf(keyFormatUser, user)
}

func FormatGroupUserRelationKey(group string) string {
	return fmt.Sprintf(keyFormatGroupUserRelation, group)
}
