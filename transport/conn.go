package transport

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
)

type Conn struct {
	net.Conn
	state int
}

func (c *Conn) handleConnection(l *Listener, sessionId string) {
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
			c.cleanUp(l, sessionId)
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
			c.cleanUp(l, sessionId)

			break
		}

		err = frame.Encode(c, frame.MessageOK)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(time.Since(now))
	}
}

func (c *Conn) cleanUp(l *Listener, sessionId string) {
	l.mu.Lock()
	delete(l.connections, sessionId)
	l.mu.Unlock()
}
