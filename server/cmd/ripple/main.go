package main

import (
	"log"

	"github.com/Phantomvv1/Ripple/session"
)

func main() {
	conn, err := session.NewSession(":42069")
	if err != nil {
		log.Println(err)
		return
	}

	err = conn.Run()
	log.Println(err)
}
