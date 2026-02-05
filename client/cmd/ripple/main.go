package main

import (
	"log"
	"net"

	"github.com/Phantomvv1/Ripple/frame"
)

func main() {
	conn, err := net.Dial("tcp", ":42069")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

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
}
