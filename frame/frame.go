package frame

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
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

func (m Message) Equals(msg Message) bool {
	if m.version != msg.version {
		return false
	}

	if m.flags != msg.flags {
		return false
	}

	if m.msgType != msg.msgType {
		return false
	}

	if m.length != msg.length {
		return false
	}

	if !bytes.Equal(m.payload, msg.payload) {
		return false
	}

	return true
}

func (m Message) String() string {
	msgType := ""
	switch m.msgType {
	case RequestMsg:
		msgType = "request message"
	case ResponseMsg:
		msgType = "response message"
	case controlMsg:
		msgType = "control message"
	}

	return fmt.Sprintf("Message v%d, flags: %s, type: %s\nlength: %d\npayload: %s", m.version, strconv.FormatInt(int64(m.flags), 2), msgType, m.length, string(m.payload))
}

// This method is used to decode a json payload into the provided value. The value must be a pointer!
func (m *Message) DecodeJSONPayload(v any) error {
	err := json.Unmarshal(m.payload, v)
	if err != nil {
		return err
	}

	return nil
}

// Returns true if the message is cachable and false if it is not
func (m Message) Cachable() bool {
	return m.flags&CacheFlag != 0
}

func ValidMsgType(msgType byte) bool {
	if msgType != RequestMsg && msgType != ResponseMsg && msgType != controlMsg {
		return false
	}

	return true
}

func ValidPayloadSize(length uint32) bool {
	if length >= uint32(MaxPayloadSize) {
		return false
	}

	return true
}

// flags:
// bytes 0 % 3 - Methods
// byte 4: is request cachable; 0 - not, 1 - cache the request
func NewMessage(payload []byte, msgType byte, flags byte) (*Message, error) {
	if !ValidMsgType(msgType) {
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

// This function gets the payload, encodes it to json and returns a new message with the json encoded payload
func NewJSONMessage[T any](payload T, msgType byte, flags byte) (*Message, error) {
	msgPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return NewMessage(msgPayload, msgType, flags)
}

func Encode(w io.Writer, m *Message) error {
	if !ValidPayloadSize(m.length) {
		return errors.New("Error: the size of the payload is too big")
	}

	buf := make([]byte, 0, m.length+9) // magics + version + flags + msgType + length

	buf = append(buf, Magic1, Magic2, m.version, m.flags, m.msgType)

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, m.length)
	buf = append(buf, lenBuf...)

	buf = append(buf, m.payload...)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}

	if n != len(buf) {
		return errors.New("Error: the number of bytes written is smaller than the intended")
	}

	return nil
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

	if !ValidMsgType(header[4]) {
		return nil, errors.New("Error: unknown message type")
	}

	msg.version = Version
	msg.flags = header[3]
	msg.msgType = header[4]
	msg.length = binary.BigEndian.Uint32(header[5:9])

	if !ValidPayloadSize(msg.length) {
		return nil, errors.New("Error: the size of the payload is too big")
	}

	payload := make([]byte, msg.length)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, err
	}

	msg.payload = payload

	return msg, nil
}
