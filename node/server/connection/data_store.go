package connection

type dataStore interface {
	PopNewData([]byte, error)
}

// type ReportObject interface {
// 	NewMessage([]byte)
// 	ConnError(string)
// }
