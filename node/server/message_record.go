package server

const (
	msg_state_default   = 0
	msg_state_buffering = 1
	msg_state_sent      = 2
)

type messageRecord struct {
	// SendState int //发送状态: 0,初始状态;1,缓存中;2,已发送. 当连接重置时,如果消息状态为1,需要将消息的状态置为初始状态0
	*Message
}

func NewMessageRecord(msg *Message) *messageRecord {
	mr := &messageRecord{
		Message: msg,
		// Receiver: receiver,
	}
	return mr
}

// func (mr *messageRecord) sendMessage(connections ConnectionList) error {

// 	c, exists := connections.Find(mr.Receiver)
// 	if exists == false { //use is offline
// 		return nil
// 	}

// 	err := c.Send(mr.MessageBytes)
// 	if err != nil {
// 		// send data next time, records after this will not use this connection
// 		// connections.Remove(mr.Receiver)
// 		return err
// 	}
// 	return nil
// 	// for receiver := range mr.Receivers {

// 	// 	c, exists := connections.Find(receiver)
// 	// 	if exists == false { //use is offline
// 	// 		// log.DebugSysF("%s has been remove from group", receiver)
// 	// 		// delete(mr.Receivers, receiver)
// 	// 		// log.TraceF("receiver %s not found", receiver)
// 	// 		continue
// 	// 	}

// 	// 	err := c.Send(mr.MessageBytes)
// 	// 	if err == nil {
// 	// 		//if message has been sent successfully, remove the receiver for duplicated sending
// 	// 		delete(mr.Receivers, receiver)
// 	// 		// log.TraceF("--> %s : %s", receiver, mr.message.String())
// 	// 	} else {
// 	// 		// send data next time, records after this will not use this connection
// 	// 		connections.Remove(receiver)
// 	// 	}
// 	// }
// }

type MessageRecordArray []*messageRecord

// func (mrl MessageRecordArray) clear() MessageRecordArray {
// 	list := MessageRecordArray{}
// 	for _, r := range mrl {
// 		if len(r.Receivers) > 0 {
// 			list = append(list, r)
// 		}
// 	}
//
// 	return list
// }

// func (mrl MessageRecordArray) sendMessage(connections connectionList) {
// 	for _, r := range mrl {
// 		r.sendMessage(connections)
// 	}
// }
