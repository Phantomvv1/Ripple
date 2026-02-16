package transport

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Phantomvv1/Ripple/frame"
)

const (
	StateInit = iota
	StateHandshake
	StateReady
)

type HandleFunc func(*Conn, *frame.Message) (*frame.Message, error)

type Conn struct {
	net.Conn
	state         int
	responseCache map[string]*frame.Message
	authEnabled   bool
	token         string
}

func newConn(conn net.Conn, authEnabled bool) *Conn {
	return &Conn{
		Conn:          conn,
		state:         StateInit,
		responseCache: make(map[string]*frame.Message),
		authEnabled:   authEnabled,
	}
}

func hashMessage(m *frame.Message) string {
	algorithm := sha256.New()
	lenght := make([]byte, 4)
	binary.BigEndian.PutUint32(lenght, m.Length())
	algorithm.Write([]byte{m.Version(), m.MsgType()})
	algorithm.Write(lenght)
	algorithm.Write(m.Payload())

	result := algorithm.Sum(nil)
	return fmt.Sprintf("%x", result)
}

func (c *Conn) handleConnection(connections map[string]*Conn, mu *sync.Mutex, sessionId string) {
	defer c.cleanUp(connections, mu, sessionId)
	now := time.Now()

	c.SetDeadline(time.Now().Add(time.Second))
	err := c.handshake()
	if err != nil {
		log.Println(err)
		c.Close()
		return
	}
	c.SetDeadline(time.Time{})

	log.Println(time.Since(now))

	if c.state != StateReady {
		log.Println("Error: the conn is not ready to work propperly")
		err := frame.Encode(c, frame.MessageClose)
		if err != nil {
			log.Println(err)
			return
		}
	}

	for {
		msg, err := frame.Decode(c)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(msg)

		if msg.Equals(*frame.MessagePing) {
			err = frame.Encode(c, frame.MessagePong)
			if err != nil {
				log.Println(err)
				return
			}

			continue
		} else if msg.Equals(*frame.MessageClose) {
			err = frame.Encode(c, frame.MessageClose)
			if err != nil {
				log.Println(err)
				c.Close()
				return
			}

			c.Close()
			return
		}

		msgHash := hashMessage(msg)
		cachable := msg.IsFlagSet(frame.CachableFlag)

		if cachable {
			if resp, ok := c.responseCache[msgHash]; ok {
				err = frame.Encode(c, resp)
				if err != nil {
					log.Println(err)
					return
				}

				continue
			}
		}

		err = frame.Encode(c, frame.MessageOK)
		if err != nil {
			log.Println(err)
			return
		}

		if cachable {
			c.responseCache[msgHash] = frame.MessageOK
		}

		fmt.Println(time.Since(now))
	}
}

func (c *Conn) handshake() error {
	c.state = StateHandshake

	err := frame.Encode(c, frame.MessageWelcome)
	if err != nil {
		return err
	}

	c.state = StateReady

	return nil
}

func (c *Conn) cleanUp(connections map[string]*Conn, mu *sync.Mutex, sessionId string) {
	mu.Lock()
	delete(connections, sessionId)
	mu.Unlock()
}
