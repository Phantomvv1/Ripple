package transport

import (
	"crypto/rand"
	"net"
	"sync"

	"github.com/Phantomvv1/Ripple/frame"
)

type Listener struct {
	listener    net.Listener
	authEnabled bool
	connections map[string]*Conn
	mu          sync.Mutex
}

func NewListener(addr string) (*Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Listener{
		listener:    listener,
		authEnabled: true,
		connections: make(map[string]*Conn),
		mu:          sync.Mutex{},
	}, nil
}

func NewListenerWithoutAuth(addr string) (*Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Listener{
		listener:    listener,
		authEnabled: false,
		connections: make(map[string]*Conn),
		mu:          sync.Mutex{},
	}, nil
}

func (l *Listener) Run() error {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			return err
		}

		sessionId := rand.Text()

		upgradedConn := newConn(conn)

		l.mu.Lock()
		l.connections[sessionId] = upgradedConn
		l.mu.Unlock()

		go upgradedConn.handleConnection(l, sessionId)
	}
}

func (l *Listener) handshake(conn *Conn) error {
	conn.state = StateHandshake

	err := frame.Encode(conn, frame.MessageWelcome)
	if err != nil {
		return err
	}

	conn.state = StateReady

	return nil
}
