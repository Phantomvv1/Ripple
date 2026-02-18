package client

import (
	"errors"
	"net"

	"github.com/Phantomvv1/Ripple/frame"
)

type ClientConn struct {
	net.Conn
	authEnabled bool
	authToken   [16]byte
}

func (c ClientConn) AuthEnabled() bool {
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

func (c *ClientConn) SendMessage(msg *frame.Message) error {
	if c.authEnabled {
		msg.UpdateAuthToken(c.authToken)
		msg.UpdateFlag(frame.AuthEnabledFlag)
	}

	return frame.Encode(c, msg)
}

func (c *ClientConn) ReceiveMessage() (*frame.Message, error) {
	msg, err := frame.Decode(c)
	if err != nil {
		return nil, err
	}

	if c.authEnabled {
		c.authToken = msg.AuthToken()
	}

	return msg, nil
}

func NewClientConn(conn net.Conn) *ClientConn {
	return &ClientConn{Conn: conn, authEnabled: false}
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
