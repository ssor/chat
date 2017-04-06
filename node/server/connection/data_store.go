package connection

type dataStore interface {
	NewDataIn([]byte) error
}

// type ReportObject interface {
// 	NewMessage([]byte)
// 	ConnError(string)
// }
