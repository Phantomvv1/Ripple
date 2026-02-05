package session

import (
	"errors"
	"log"
	"net"
)

type Router struct {
	connection  *net.TCPConn
	authEnabled bool
}

func NewRouter() *Router {
	return &Router{
		connection:  nil,
		authEnabled: true,
	}
}

func NewRouterWithoutAuth() *Router {
	return &Router{
		connection:  nil,
		authEnabled: false,
	}
}

func (r *Router) Run(addr *net.TCPAddr) error {
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Error: failed to establish connection")
			innerErr := conn.Close()
			if !errors.Is(innerErr, net.ErrClosed) {
				return innerErr
			}
		}

		r.connection = conn
	}
}
