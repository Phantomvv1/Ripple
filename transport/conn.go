package transport

import (
	"crypto/rand"
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

type HandleFunc func(*frame.Message) (*frame.Message, error)

type Conn struct {
	net.Conn
	state         int
	responseCache map[string]*frame.Message
	token         [16]byte
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
	algorithm.Write([]byte{m.Version(), m.MsgType()})
	algorithm.Write(lenght)
	algorithm.Write(m.Payload())

	result := algorithm.Sum(nil)
	return fmt.Sprintf("%x", result)
}

func (c *Conn) handleConnection(connections map[string]*Conn, mu *sync.Mutex, sessionId string, operations map[int]HandleFunc) {
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

		if frame.AuthEnabled && msg.AuthToken() == c.token {
			token, err := makeAuthToken()
			if err != nil {
				log.Println(err)
				return
			}

			c.token = *token
		} else {
			log.Println(c.token)
			log.Println("Error: wrong token provided")
			return
		}

		if msg.Equals(*frame.MessagePing) {
			err = frame.Encode(c, frame.MessagePong)
			if err != nil {
				log.Println(err)
				return
			}

			continue
		}

		if msg.Equals(*frame.MessageClose) {
			err = frame.Encode(c, frame.MessageClose)
			if err != nil {
				log.Println(err)
				return
			}

			return
		}

		msgHash := hashMessage(msg)
		cachable := msg.IsFlagSet(frame.CachableFlag)

		if cachable {
			if resp, ok := c.responseCache[msgHash]; ok {
				if frame.AuthEnabled {
					resp.UpdateAuthToken(c.token)
				}

				err = frame.Encode(c, resp)
				if err != nil {
					log.Println(err)
					return
				}

				continue
			}
		}

		if frame.AuthEnabled {
			frame.MessageOK.UpdateAuthToken(c.token)
		}

		handler, ok := operations[int(msg.OperationId())]
		if !ok {
			errMsg := "Error: there is no operation for this operation id"
			msg, err := frame.NewMessage([]byte(errMsg), frame.ResponseMsg, 0)
			if err != nil {
				log.Println(err)
				return
			}

			err = frame.Encode(c, msg)
			if err != nil {
				log.Println(err)
				return
			}
		}

		resp, err := handler(msg)
		if err != nil {
			log.Println(err)
		}

		if frame.AuthEnabled {
			resp.UpdateAuthToken(c.token)
		}

		err = frame.Encode(c, resp)
		if err != nil {
			log.Println(err)
			return
		}

		if cachable {
			c.responseCache[msgHash] = resp
		}

		fmt.Println(time.Since(now))
	}
}

func (c *Conn) handshake() error {
	c.state = StateHandshake

	token, err := makeAuthToken()
	if err != nil {
		return err
	}

	frame.MessageWelcome.UpdateAuthToken(*token)
	c.token = *token

	err = frame.Encode(c, frame.MessageWelcome)
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

	c.Close()
}

func makeAuthToken() (*[16]byte, error) {
	result := make([]byte, 16)
	_, err := rand.Read(result)
	if err != nil {
		return nil, err
	}

	r := [16]byte(result)
	return &r, nil
}
