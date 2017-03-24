package server

type messageRecord struct {
	*Message
}

func newMessageRecord(msg *Message) *messageRecord {
	mr := &messageRecord{
		Message: msg,
		// Receiver: receiver,
	}
	return mr
}

type messageRecordArray []*messageRecord
