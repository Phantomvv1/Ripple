package transport

import (
	"crypto/hmac"
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
	secret        [32]byte
	writeChan     chan *frame.Message
	writeErrMap   map[uint32]chan error // seqNumber -> err
	mu            sync.Mutex
}

func newConn(conn net.Conn) (*Conn, error) {
	c := &Conn{
		Conn:          conn,
		state:         StateInit,
		responseCache: make(map[string]*frame.Message),
		writeChan:     make(chan *frame.Message),
		writeErrMap:   make(map[uint32]chan error),
		mu:            sync.Mutex{},
	}

	c.SetDeadline(time.Now().Add(time.Second))
	err := c.handshake()
	if err != nil {
		log.Println(err)
		c.Close()
		return nil, err
	}
	c.SetDeadline(time.Time{})

	go c.writeMessages()

	return c, nil
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
	now := time.Now().UTC()

	if c.state != StateReady {
		log.Println("Error: the conn is not ready to work propperly")
		err := c.Send(frame.MessageClose)
		if err != nil {
			log.Println(err)
			return
		}
	}

	for {
		receivedMsg, err := frame.Decode(c)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(receivedMsg)
		fmt.Println()

		go func(receivedMsg *frame.Message) {
			if receivedMsg.Equals(*frame.MessagePing) {
				pingMsg := frame.MessagePing.Clone().UpdateSequenceNumber(receivedMsg.SequenceNumber())
				encodeErrChan := c.SendConcurrentMsg(pingMsg)
				err = <-encodeErrChan
				if err != nil {
					log.Println(err)
					return
				}
			}

			if receivedMsg.Equals(*frame.MessageClose) {
				closeMsg := frame.MessageClose.Clone().UpdateSequenceNumber(receivedMsg.SequenceNumber())
				encodeErrChan := c.SendConcurrentMsg(closeMsg)
				err = <-encodeErrChan
				if err != nil {
					log.Println(err)
					return
				}

				fmt.Println(time.Since(now))
				return
			}

			if frame.AuthEnabled {
				sentToken := receivedMsg.AuthToken()
				checkToken := c.makeAuthToken(receivedMsg.SequenceNumber(), receivedMsg.Payload())
				if !hmac.Equal(sentToken[:], checkToken[:]) {
					log.Println("Tampared message")
					return
				}
			}

			msgHash := hashMessage(receivedMsg)
			cachable := receivedMsg.IsFlagSet(frame.CachableFlag)

			if cachable {
				if resp, ok := c.responseCache[msgHash]; ok {
					encodeErrChan := c.SendConcurrentMsg(resp)
					err = <-encodeErrChan
					if err != nil {
						log.Println(err)
						return
					}

					fmt.Println(time.Since(now))
				}
			}

			handler, ok := operations[int(receivedMsg.OperationId())]
			if !ok {
				errMsg := "Error: there is no operation for this operation id"
				msg, err := frame.NewMessage([]byte(errMsg), frame.ResponseMsg, 0)
				if err != nil {
					log.Println(err)
					return
				}

				err = c.Send(msg)
				if err != nil {
					log.Println(err)
					return
				}
			}

			resp, err := handler(receivedMsg)
			if err != nil {
				log.Println(err)
			}

			resp = resp.Clone()
			resp.UpdateSequenceNumber(receivedMsg.SequenceNumber())

			encodeErrChan := c.SendConcurrentMsg(resp)
			err = <-encodeErrChan
			if err != nil {
				log.Println(err)
				return
			}

			if cachable {
				c.responseCache[msgHash] = resp
			}

			fmt.Println(time.Since(now))
		}(receivedMsg)
	}
}

func (c *Conn) handshake() error {
	c.state = StateHandshake

	secret, err := c.makeSessionSecret()
	if err != nil {
		return err
	}

	c.secret = *secret

	welcomeMsg := frame.MessageWelcome.Clone().UpdateAuthToken(*secret)

	err = c.Send(welcomeMsg)
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

func (c *Conn) makeSessionSecret() (*[32]byte, error) {
	result := make([]byte, 32)
	_, err := rand.Read(result)
	if err != nil {
		return nil, err
	}

	r := [32]byte(result)
	return &r, nil
}

func (c *Conn) Send(msg *frame.Message) error {
	return frame.Encode(c, msg, msg.SequenceNumber())
}

func (c *Conn) SendConcurrentMsg(msg *frame.Message) chan error {
	encodeErrChan := make(chan error)

	c.mu.Lock()
	c.writeErrMap[msg.SequenceNumber()] = encodeErrChan
	c.mu.Unlock()

	c.writeChan <- msg

	return encodeErrChan
}

func (c *Conn) makeAuthToken(seqNumber uint32, payload []byte) [32]byte {
	algorithm := hmac.New(sha256.New, c.secret[:])
	sequenceNumberSlice := make([]byte, 4)
	binary.BigEndian.PutUint32(sequenceNumberSlice, seqNumber)
	algorithm.Write(sequenceNumberSlice)
	algorithm.Write(payload)

	token := algorithm.Sum(nil)
	return [32]byte(token)
}

func (c *Conn) writeMessages() {
	for {
		msg := <-c.writeChan

		err := c.Send(msg)

		c.mu.Lock()
		ch := c.writeErrMap[msg.SequenceNumber()]
		c.mu.Unlock()

		ch <- err
	}
}
