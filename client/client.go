package client

import (
	"errors"
	"net"
	"sync"

	"github.com/Phantomvv1/Ripple/frame"
)

type ClientConn struct {
	net.Conn
	authEnabled     bool
	authToken       [16]byte
	sequenceNumber  uint32
	pendingMessages map[uint32]*frame.Message
	mu              sync.Mutex
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
		c.authToken = msg.AuthToken()
	}

	return nil
}

func (c *ClientConn) SendMessage(msg *frame.Message) (uint32, error) {
	if c.authEnabled {
		msg.UpdateAuthToken(c.authToken)
		msg.UpdateFlag(frame.AuthEnabledFlag)
	}

	var err error
	if !msg.Equals(*frame.MessageClose) && !msg.Equals(*frame.MessagePing) {
		for {
			if _, ok := c.pendingMessages[c.sequenceNumber]; ok {
				c.sequenceNumber++
				continue
			}

			err = frame.Encode(c, msg, c.sequenceNumber)

			c.mu.Lock()
			c.pendingMessages[c.sequenceNumber] = nil
			c.mu.Unlock()

			c.sequenceNumber++

			return c.sequenceNumber - 1, err
		}

	}

	return 0, frame.Encode(c, msg, 0)
}

func (c *ClientConn) ReceiveMessage(sequenceNumber uint32) (*frame.Message, error) {
	msg, err := frame.Decode(c)
	if err != nil {
		return nil, err
	}

	if msg.SequenceNumber() != sequenceNumber {
		c.pendingMessages[msg.SequenceNumber()] = msg

		res := make(chan *frame.Message, 1)
		c.listenForResponse(res, sequenceNumber)

		msg = <-res
	}

	if c.authEnabled {
		c.authToken = msg.AuthToken()
	}

	delete(c.pendingMessages, msg.SequenceNumber())

	return msg, err
}

func NewClientConn(conn net.Conn) *ClientConn {
	return &ClientConn{
		Conn:            conn,
		authEnabled:     false,
		sequenceNumber:  0,
		pendingMessages: make(map[uint32]*frame.Message),
		mu:              sync.Mutex{},
	}
}

// This function dials the server on the given port. The format of the port parameter is as follows: ":8080"
func Dial(port string) (*ClientConn, error) {
	conn, err := net.Dial("tcp", port)
	if err != nil {
		return nil, err
	}

	clientConn := NewClientConn(conn)

	err = clientConn.handshake()
	if err != nil {
		return nil, err
	}

	return clientConn, nil
}

func (c *ClientConn) listenForResponse(res chan<- *frame.Message, sequenceNumber uint32) {
	for {
		if resp, ok := c.pendingMessages[sequenceNumber]; ok {
			res <- resp
			return
		}
	}
}
