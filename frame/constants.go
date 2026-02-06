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
	MessageHello, _ = NewMessage([]byte{0}, ControlMsg, 0)
	//This will be used for the handshake
	MessageWelcome, _ = NewMessage([]byte{1}, ControlMsg, 0)
	//This will be used for the handshake
	MessageReject, _ = NewMessage([]byte{2}, ControlMsg, 0)

	MessageClose, _ = NewMessage([]byte{3}, ControlMsg, 0)
	MessagePing, _  = NewMessage([]byte("Ping"), ControlMsg, 0)
	MessagePong, _  = NewMessage([]byte("Pong"), ControlMsg, 0)
)
