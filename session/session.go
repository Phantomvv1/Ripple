package session

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Phantomvv1/Ripple/frame"
)

const (
	StateInit = iota
	StateHandshake
	StateReady
	StateClosed
)

type Conn struct {
	listener    net.Listener
	state       int
	authEnabled bool
}

func NewSession(addr string) (*Conn, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Conn{
		listener:    listener,
		state:       StateInit,
		authEnabled: true,
	}, nil
}

func NewSessionWithoutAuth(addr string) (*Conn, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Conn{
		listener:    listener,
		state:       StateInit,
		authEnabled: false,
	}, nil
}

func (c *Conn) Run() error {
	for {
		conn, err := c.listener.Accept()

		now := time.Now()
		err = c.handshake(conn)
		if err != nil {
			return err
		}
		log.Println(time.Since(now))

		msg, err := frame.Decode(conn)
		if err != nil {
			return err
		}

		fmt.Println(msg)

		if msg.Equals(*frame.MessagePing) {
			err = frame.Encode(conn, frame.MessagePong)
			if err != nil {
				return err
			}

			continue
		}

		err = frame.Encode(conn, frame.MessageOK)
		if err != nil {
			return err
		}
	}
}

func (c *Conn) handshake(conn net.Conn) error {
	c.state = StateHandshake

	err := frame.Encode(conn, frame.MessageWelcome)
	if err != nil {
		return err
	}

	c.state = StateReady

	return nil
}
