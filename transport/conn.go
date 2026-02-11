package transport

import (
	"crypto/sha256"
	"encoding/binary"
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
)

type Conn struct {
	net.Conn
	state         int
	responseCache map[string]*frame.Message
}

func newConn(conn net.Conn) *Conn {
	return &Conn{
		Conn:          conn,
		state:         StateInit,
		responseCache: make(map[string]*frame.Message),
	}
}

func hashMessage(m *frame.Message) string {
	algorithm := sha256.New()
	lenght := make([]byte, 4)
	binary.BigEndian.PutUint32(lenght, m.Length())
	algorithm.Write([]byte{m.Version(), m.Flags(), m.MsgType()})
	algorithm.Write(lenght)
	algorithm.Write(m.Payload())

	result := algorithm.Sum(nil)
	return fmt.Sprintf("%x", result)
}

func (c *Conn) handleConnection(l *Listener, sessionId string) {
	defer c.cleanUp(l, sessionId)
	now := time.Now()

	c.SetDeadline(time.Now().Add(time.Second))
	err := l.handshake(c)
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

		if msg.Cachable() {
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

		if msg.Cachable() {
			c.responseCache[msgHash] = frame.MessageOK
		}

		fmt.Println(time.Since(now))
	}
}

func (c *Conn) cleanUp(l *Listener, sessionId string) {
	l.mu.Lock()
	delete(l.connections, sessionId)
	l.mu.Unlock()
}
