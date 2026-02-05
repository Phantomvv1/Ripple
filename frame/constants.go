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
	ResponseMsgOK, _ = NewMessage(nil, ResponseMsg, 0)
)
