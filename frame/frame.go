package frame

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type Message struct {
	version        byte
	flags          byte
	msgType        byte
	operationId    uint16
	authToken      [16]byte
	sequenceNumber uint32
	length         uint32
	payload        []byte
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

func (m Message) AuthToken() [16]byte {
	return m.authToken
}

func (m Message) OperationId() uint16 {
	return m.operationId
}

func (m *Message) SequenceNumber() uint32 {
	return m.sequenceNumber
}

func (m *Message) UpdateAuthToken(token [16]byte) {
	m.authToken = token
}

func (m Message) Equals(msg Message) bool {
	if m.version != msg.version {
		return false
	}

	if m.msgType != msg.msgType {
		return false
	}

	if m.msgType != controlMsg {
		if m.flags != msg.flags {
			return false
		}
	}

	if m.msgType == RequestMsg {
		if m.operationId != msg.operationId {
			return false
		}
	}

	if m.sequenceNumber != msg.sequenceNumber {
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

	if m.IsFlagSet(AuthEnabledFlag) {
		if m.msgType == RequestMsg {
			return fmt.Sprintf("Message v%d, flags: %b, type: %s\ntoken: %v\noperationId: %d\nsequence number: %d\nlength: %d\npayload: %s",
				m.version, m.flags, msgType, m.authToken, m.operationId, m.sequenceNumber, m.length, string(m.payload))
		} else {
			return fmt.Sprintf("Message v%d, flags: %b, type: %s\ntoken: %v\nsequence number: %d\nlength: %d\npayload: %s",
				m.version, m.flags, msgType, m.authToken, m.sequenceNumber, m.length, string(m.payload))
		}
	} else {
		if m.msgType == RequestMsg {
			return fmt.Sprintf("Message v%d, flags: %b, type: %s\noperationId: %d\nsequence number: %d\nlength: %d\npayload: %s",
				m.version, m.flags, msgType, m.operationId, m.sequenceNumber, m.length, string(m.payload))
		} else {
			return fmt.Sprintf("Message v%d, flags: %b, type: %s\nsequence number: %d\nlength: %d\npayload: %s",
				m.version, m.flags, msgType, m.sequenceNumber, m.length, string(m.payload))
		}
	}
}

// This method is used to decode a json payload into the provided value. The value must be a pointer!
func (m *Message) DecodeJSONPayload(v any) error {
	return json.Unmarshal(m.payload, v)
}

// Returns true if the flag in the message is 1 and false if it is not
func (m Message) IsFlagSet(flag byte) bool {
	return m.flags&flag != 0
}

func (m *Message) UpdateFlag(flag byte) {
	m.flags |= flag
}

func (m *Message) UpdateSequenceNumber(sequenceNumber uint32) {
	m.sequenceNumber = sequenceNumber
}

func (m *Message) CompressPayload() error {
	return nil
}

func (m *Message) DecompressPayload() error {
	return nil
}

func (m *Message) EncryptPayload() error {
	return nil
}

func (m *Message) DecryptPayload() error {
	return nil
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

// The new message function creates a new message with the given payload, msgType and flags.
// Depending on which flags are set the message is automatically compressed and/or encrypted
// When creating a request message the user must always provide an operationId
func NewMessage(payload []byte, msgType byte, flags byte, operationId ...uint16) (*Message, error) {
	if !ValidMsgType(msgType) {
		return nil, errors.New("Error: unknown message type")
	}

	if AuthEnabled && msgType != controlMsg {
		flags |= AuthEnabledFlag
	}

	msg := &Message{
		version: Version,
		flags:   flags,
		msgType: msgType,
		length:  uint32(len(payload)),
		payload: payload,
	}

	if msgType == RequestMsg {
		if operationId != nil {
			msg.operationId = operationId[0]
		} else {
			return nil, errors.New("Error: when creating a request message you must always provide an operationId")
		}

	}
	payloadChange := false
	if msg.IsFlagSet(CompressedPayloadFlag) {
		err := msg.CompressPayload()
		if err != nil {
			return nil, err
		}

		payloadChange = true
	}

	if msg.IsFlagSet(EncryptedPayloadFlag) {
		err := msg.EncryptPayload()
		if err != nil {
			return nil, err
		}

		payloadChange = true
	}

	if payloadChange {
		msg.length = uint32(len(msg.payload))
	}

	return msg, nil
}

// The new json message function creates a new message with the given payload encoded into a json format, msgType and flags.
// Depending on which flags are set the message is automatically compressed and/or encrypted
func NewJSONMessage[T any](payload T, msgType byte, flags byte, operationId ...uint16) (*Message, error) {
	msgPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return NewMessage(msgPayload, msgType, flags, operationId...)
}

func Encode(w io.Writer, m *Message, sequenceNumber uint32) error {
	if !ValidPayloadSize(m.length) {
		return errors.New("Error: the size of the payload is too big")
	}

	isRequestMsg := m.msgType == RequestMsg

	headerSize := 2 + 1 + 1 + 1 + 2 + 4 // Magics (2) + Version(1) + Flags(1) + MsgType(1) + SequenceNumber(2) + Length(4)
	if AuthEnabled {
		headerSize += 16 // AuthToken(16)
	}

	if isRequestMsg {
		headerSize += 2 // OperationId(2)
	}

	buf := make([]byte, 0, headerSize+int(m.length))

	buf = append(buf, Magic1, Magic2, m.version, m.flags, m.msgType)

	buf = binary.BigEndian.AppendUint32(buf, sequenceNumber)

	if isRequestMsg {
		operationId := make([]byte, 2)
		binary.BigEndian.PutUint16(operationId, m.operationId)

		buf = append(buf, operationId...)
	}

	if m.IsFlagSet(AuthEnabledFlag) {
		buf = append(buf, m.authToken[:]...)
	}

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
	msg.sequenceNumber = binary.BigEndian.Uint32(header[5:9])

	if msg.msgType == RequestMsg {
		operationId := make([]byte, 2)
		_, err = io.ReadFull(r, operationId)
		if err != nil {
			return nil, err
		}

		msg.operationId = binary.BigEndian.Uint16(operationId)
	}

	if msg.IsFlagSet(AuthEnabledFlag) {
		token := make([]byte, 16)
		_, err = io.ReadFull(r, token)
		if err != nil {
			return nil, err
		}

		msg.authToken = [16]byte(token)
	}

	length := make([]byte, 4)
	_, err = io.ReadFull(r, length)
	if err != nil {
		return nil, err
	}

	msg.length = binary.BigEndian.Uint32(length)

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
