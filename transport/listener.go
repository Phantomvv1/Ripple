package transport

import (
	"crypto/rand"
	"errors"
	"net"
	"reflect"
	"runtime"
	"sync"

	"github.com/Phantomvv1/Ripple/frame"
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

	frame.AuthEnabled = true

	return &Listener{
		listener:    listener,
		authEnabled: frame.AuthEnabled,
		connections: make(map[string]*Conn),
		mu:          sync.Mutex{},
		operations:  make(map[int]HandleFunc),
	}, nil
}

func NewListenerWithoutAuth(addr string) (*Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	frame.AuthEnabled = false

	return &Listener{
		listener:    listener,
		authEnabled: frame.AuthEnabled,
		connections: make(map[string]*Conn),
		mu:          sync.Mutex{},
	}, nil
}

func (l *Listener) Run() error {
	if len(l.operations) == 0 {
		return errors.New("Error: you can't run the listener when you have 0 operations to do")
	}

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

		go upgradedConn.handleConnection(l.connections, &l.mu, sessionId, l.operations)
	}
}

func (l *Listener) AddOperation(operationId int, handleFunc HandleFunc) {
	if handler, ok := l.operations[operationId]; ok {
		funcName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		panic("Error: adding an operation with an id that already exists for the function " + funcName)
	}

	l.operations[operationId] = handleFunc
}
