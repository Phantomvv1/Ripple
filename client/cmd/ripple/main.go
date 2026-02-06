package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Phantomvv1/Ripple/frame"
)

func main() {
	conn, err := net.Dial("tcp", ":42069")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	now := time.Now()

	err = frame.Encode(conn, frame.MessageHello)
	if err != nil {
		log.Println(err)
		return
	}

	msg, err := frame.Decode(conn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(msg)

	m, err := frame.NewMessage([]byte("Test"), frame.RequestMsg, 0)
	if err != nil {
		log.Println(err)
		return
	}

	err = frame.Encode(conn, m)
	if err != nil {
		log.Println(err)
		return
	}

	msg, err = frame.Decode(conn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(msg)

	err = frame.Encode(conn, frame.MessagePing)
	if err != nil {
		log.Println(err)
		return
	}

	msg, err = frame.Decode(conn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(msg)
	fmt.Println(time.Since(now))
}
