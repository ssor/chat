package communication

var (
	// ProtoUnknown unknown
	ProtoUnknown = -1
	// protoLogin  = 0
	// protoLogout = 1

	// ProtoText text
	ProtoText = 2
	// ProtoImage image upload
	ProtoImage = 3
	// ProtoAudio audio upload
	ProtoAudio = 4
	// ProtoShare share some thing
	ProtoShare = 5
	// ProtoReply reply from client
	ProtoReply = 6
	// ProtoBot bot chat
	ProtoBot = 888
	// ProtoCloseLoginOnOtherDevice server close conn and do not want client to reconnect in short time
	ProtoCloseLoginOnOtherDevice = "loginOnOtherDevice"
)
