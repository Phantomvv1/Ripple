package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/Phantomvv1/Ripple/frame"
)

type response struct {
	msg *frame.Message
	err error
}

type ClientConn struct {
	net.Conn
	authEnabled       bool
	secret            [32]byte
	sequenceNumber    uint32
	pendingMessages   map[uint32]response
	muSeqNum          sync.Mutex
	muPendingMessages sync.Mutex
	messageReceived   chan struct{}
	messageSent       chan uint32
	sendMessage       chan *frame.Message
}

func NewClientConn(conn net.Conn) (*ClientConn, error) {
	cl := &ClientConn{
		Conn:              conn,
		authEnabled:       false,
		sequenceNumber:    0,
		pendingMessages:   make(map[uint32]response),
		muSeqNum:          sync.Mutex{},
		muPendingMessages: sync.Mutex{},
		messageReceived:   make(chan struct{}),
		messageSent:       make(chan uint32),
		sendMessage:       make(chan *frame.Message),
	}

	err := cl.handshake()
	if err != nil {
		return nil, err
	}

	go cl.readResponses()
	go cl.writeMessages()

	return cl, nil
}

func (c *ClientConn) AuthEnabled() bool {
	return c.authEnabled
}

func (c *ClientConn) handshake() error {
	msg, err := frame.Decode(c)
	if err != nil {
		return err
	}

	if !msg.Equals(*frame.MessageWelcome) {
		return errors.New("Error: the server has rejected the connection")
	}

	if msg.IsFlagSet(frame.AuthEnabledFlag) {
		c.authEnabled = true
		c.secret = msg.AuthToken()
	}

	return nil
}

func (c *ClientConn) SendMessage(msg *frame.Message) (*frame.Message, error) {
	c.muSeqNum.Lock()

	seq := c.sequenceNumber
	msg.UpdateSequenceNumber(c.sequenceNumber)
	c.sequenceNumber++

	c.muSeqNum.Unlock()

	c.sendMessage <- msg

	// Wait until message was sent
	for {
		seqNumOfSentMsg := <-c.messageSent
		if seqNumOfSentMsg == seq {
			break
		}
	}

	//Check for error about encoding the message and delete the information
	c.muPendingMessages.Lock()
	res, _ := c.pendingMessages[seq]

	if res.msg == nil {
		if res.err != nil {
			return nil, res.err
		}

		delete(c.pendingMessages, seq)
		c.muPendingMessages.Unlock()
	} else {
		c.muPendingMessages.Unlock()
		return res.msg, res.err
	}

	for {
		<-c.messageReceived

		resp, ok := c.pendingMessages[seq]
		if ok {
			return resp.msg, resp.err
		}
	}
}

// This function dials the server on the given port. The format of the port parameter is as follows: ":8080"
func Dial(port string) (*ClientConn, error) {
	conn, err := net.Dial("tcp", port)
	if err != nil {
		return nil, err
	}

	return NewClientConn(conn)
}

func (c *ClientConn) readResponses() {
	for {
		msg, err := frame.Decode(c)
		if err != nil {
			log.Println(err)
		}

		c.muPendingMessages.Lock()
		c.pendingMessages[msg.SequenceNumber()] = response{msg: msg, err: err}
		c.muPendingMessages.Unlock()

		c.messageReceived <- struct{}{}
	}
}

func (c *ClientConn) writeMessages() {
	for {
		msg := <-c.sendMessage

		if c.authEnabled {
			msg.UpdateAuthToken(c.makeAuthToken(msg.SequenceNumber(), msg.Payload()))
			msg.UpdateFlag(frame.AuthEnabledFlag)
		}

		err := frame.Encode(c, msg, msg.SequenceNumber())
		if err != nil {
			log.Println(err)
		}

		c.muPendingMessages.Lock()
		c.pendingMessages[msg.SequenceNumber()] = response{msg: nil, err: err}
		c.muPendingMessages.Unlock()

		c.messageSent <- msg.SequenceNumber()
	}
}

func (c *ClientConn) makeAuthToken(seqNumber uint32, payload []byte) [32]byte {
	algorithm := hmac.New(sha256.New, c.secret[:])
	sequenceNumberSlice := make([]byte, 4)
	binary.BigEndian.PutUint32(sequenceNumberSlice, seqNumber)
	algorithm.Write(sequenceNumberSlice)
	algorithm.Write(payload)

	token := algorithm.Sum(nil)
	return [32]byte(token)
}
