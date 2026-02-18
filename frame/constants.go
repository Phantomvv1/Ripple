package frame

const (
	controlMsg byte = iota
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
	MessageHello, _ = NewMessage([]byte{'H'}, controlMsg, CachableFlag)
	//This will be used for the handshake
	MessageWelcome, _ = NewMessage([]byte{'W'}, controlMsg, 0)
	//This will be used for the handshake
	MessageReject, _ = NewMessage([]byte{'R'}, controlMsg, 0)

	MessageClose, _ = NewMessage([]byte{'C'}, controlMsg, 0)
	MessagePing, _  = NewMessage([]byte("Ping"), controlMsg, 0)
	MessagePong, _  = NewMessage([]byte("Pong"), controlMsg, 0)
)

// This variable determines if the auth of the server is enabled.
// By default auth is enabled and the variable is set to true.
// If you want to disable auth you can set this variable to false.
// Another way of choosing to use auth or not is to call one of the functions that create a new listtener.
// They automatically tweek the variable as needed.
var AuthEnabled = true

const (
	AuthEnabledFlag       = 1 << 2
	CachableFlag          = 1 << 3
	CompressedPayloadFlag = 1 << 4
	EncryptedPayloadFlag  = 1 << 5
)

func init() {
	if AuthEnabled {
		MessageOK.flags |= AuthEnabledFlag
		MessageWelcome.flags |= AuthEnabledFlag
	}
}
