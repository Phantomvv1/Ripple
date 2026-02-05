package main

import (
	"fmt"
	"log"
	"net"
	"net/netip"

	"github.com/Phantomvv1/Ripple/frame"
)

func handleConn(conn net.Conn) {
	defer conn.Close()

	for {
		msg, err := frame.Decode(conn)
		if err != nil {
			log.Println("connection closed:", err)
			return
		}

		fmt.Println(msg)

		err = frame.Encode(conn, frame.ResponseMsgOK)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	listener, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.IPv4Unspecified(), 42069)))
	if err != nil {
		log.Println(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}
