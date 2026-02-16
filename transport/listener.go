package transport

import (
	"crypto/rand"
	"net"
	"sync"
)

type Listener struct {
	listener    net.Listener
	authEnabled bool
	operations  map[int]HandleFunc
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

		upgradedConn := newConn(conn, l.authEnabled)

		l.mu.Lock()
		l.connections[sessionId] = upgradedConn
		l.mu.Unlock()

		go upgradedConn.handleConnection(l.connections, &l.mu, sessionId)
	}
}
