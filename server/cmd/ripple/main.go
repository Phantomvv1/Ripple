package main

import (
	"log"

	"github.com/Phantomvv1/Ripple/transport"
)

func main() {
	conn, err := transport.NewListener(":42069")
	if err != nil {
		log.Println(err)
		return
	}

	err = conn.Run()
	log.Println(err)
}
