package user

type replyContent []byte

type replyContentList []replyContent

func (rcl replyContentList) Head() (replyContent, replyContentList) {
	if len(rcl) <= 0 {
		return nil, rcl
	}
	return rcl[0], rcl[1:]
}

func (rcl replyContentList) append(replys ...replyContent) replyContentList {
	return append(rcl, replys...)
}
