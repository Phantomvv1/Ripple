package frame

const (
	ControlMsg byte = iota
	RequestMsg
	ResponseMsg
)

const (
	Version byte = 1

	Magic1 byte = 'R'
	Magic2 byte = 'W'
)

var MaxPayloadSize = 1 << 21

var (
	MessageOK, _ = NewMessage(nil, ResponseMsg, 0)

	//This will be used for the handshake
	MessageHello, _ = NewMessage([]byte("Hello"), ResponseMsg, 0)
	//This will be used for the handshake
	MessageWelcome, _ = NewMessage([]byte("Welcome"), ResponseMsg, 0)
	//This will be used for the handshake
	MessageReject, _ = NewMessage([]byte("Reject"), ResponseMsg, 0)

	MessageClose, _ = NewMessage([]byte("Close"), ResponseMsg, 0)
	MessagePing, _  = NewMessage([]byte("Ping"), ResponseMsg, 0)
	MessagePong, _  = NewMessage([]byte("Pong"), ResponseMsg, 0)
)
