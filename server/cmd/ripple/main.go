package main

import (
	"log"

	"github.com/Phantomvv1/Ripple/frame"
	"github.com/Phantomvv1/Ripple/transport"
)

func testOperation(msg *frame.Message) (*frame.Message, error) {
	return frame.MessageOK, nil
}

func main() {
	conn, err := transport.NewListener(":42069")
	if err != nil {
		log.Println(err)
		return
	}

	conn.AddOperation(0, testOperation)

	log.Fatal(conn.Run())
}
