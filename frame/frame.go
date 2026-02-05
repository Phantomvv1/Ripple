package frame

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Message struct {
	version byte
	flags   byte
	msgType byte
	length  uint32
	payload []byte
}

func (m Message) Version() byte {
	return m.version
}

func (m Message) Flags() byte {
	return m.flags
}

func (m Message) MsgType() byte {
	return m.msgType
}

func (m Message) Length() uint32 {
	return m.length
}

func (m Message) Payload() []byte {
	return m.payload
}

func (m Message) String() string {
	msgType := ""
	switch m.msgType {
	case RequestMsg:
		msgType = "request message"
	case ResponseMsg:
		msgType = "response message"
	case ControlMsg:
		msgType = "control message"
	}
	return fmt.Sprintf("Message v%d, flags: %s, type: %s\nlength: %d\npayload: %s", m.version, strconv.FormatInt(int64(m.flags), 2), msgType, m.length, string(m.payload))
}

func NewMessage(payload []byte, msgType byte, flags byte) (*Message, error) {
	if msgType != RequestMsg && msgType != ResponseMsg && msgType != ControlMsg {
		return nil, errors.New("Error: unknown message type")
	}

	return &Message{
		version: Version,
		flags:   flags,
		msgType: msgType,
		length:  uint32(len(payload)),
		payload: payload,
	}, nil
}

func Encode(m *Message) ([]byte, error) {
	if m.length >= uint32(MaxPayloadSize) {
		return nil, errors.New("Error: the size of the payload is too big")
	}

	buf := make([]byte, 0, m.length+9) // magics + version + flags + msgType + length

	buf = append(buf, Magic1)
	buf = append(buf, Magic2)

	buf = append(buf, m.version)
	buf = append(buf, m.flags)
	buf = append(buf, m.msgType)

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, m.length)

	buf = append(buf, lenBuf...)

	buf = append(buf, m.payload...)

	return buf, nil
}

func Decode(r io.Reader) (*Message, error) {
	msg := &Message{}
	header := make([]byte, 9)

	_, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}

	if header[0] != Magic1 {
		return nil, errors.New("Error: missing magic 1")
	}

	if header[1] != Magic2 {
		return nil, errors.New("Error: missing magic 2")
	}

	if header[2] != Version {
		return nil, errors.New("Error: unsupported version of the protocol")
	}

	msg.version = Version
	msg.flags = header[3]
	msg.msgType = header[4]
	msg.length = binary.BigEndian.Uint32(header[5:8])

	payload := make([]byte, msg.length)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, err
	}

	msg.payload = payload

	return msg, nil
}
