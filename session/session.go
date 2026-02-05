package session

import (
	"errors"
	"fmt"
	"net"

	"github.com/Phantomvv1/Ripple/frame"
)

const (
	StateInit = iota
	StateHandshake
	StateReady
	StateClosed
)

type Conn struct {
	conn        net.Conn
	state       int
	authEnabled bool
}

func NewConn(addr string) (*Conn, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := listener.Accept()
	if err != nil {
		return nil, err
	}

	return &Conn{
		conn:        conn,
		state:       StateInit,
		authEnabled: true,
	}, nil
}

func NewConnWithoutAuth(addr string) (*Conn, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := listener.Accept()
	if err != nil {
		return nil, err
	}

	return &Conn{
		conn:        conn,
		state:       StateInit,
		authEnabled: false,
	}, nil
}

func (c *Conn) Run() error {
	c.handshake()
	for {
		msg, err := frame.Decode(c.conn)
		if err != nil {
			return err
		}

		fmt.Println(msg)

		err = frame.Encode(c.conn, frame.MessageOK)
		if err != nil {
			return err
		}
	}
}

func (c *Conn) handshake() error {
	msg, err := frame.Decode(c.conn)
	if err != nil {
		return err
	}

	if !msg.Equals(*frame.MessageHello) {
		return errors.New("Error: failed handshake - no hello message")
	}

	err = frame.Encode(c.conn, frame.MessageWelcome)
	if err != nil {
		return err
	}

	return nil
}
