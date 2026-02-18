package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Phantomvv1/Ripple/client"
	"github.com/Phantomvv1/Ripple/frame"
)

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(3)
	// go DialAndTest(wg)
	// go DialAndTest(wg)
	DialAndTest(wg)

	// wg.Wait()
}

func DialAndTest(wg *sync.WaitGroup) {
	defer wg.Done()

	now := time.Now()

	conn, err := client.Dial(":42069")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	cacheCheck := time.Now()

	msg, err := SendHelloMsg(conn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(msg)

	msg, err = SendHelloMsg(conn) // 2 times in order to check cache speed
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(msg)
	fmt.Println("Cache time: ", time.Since(cacheCheck))

	err = conn.SendMessage(frame.MessageClose)
	if err != nil {
		log.Println(err)
		return
	}

	msg, err = conn.ReceiveMessage()
	if err != nil {
		log.Println(err)
		return
	}

	if !msg.Equals(*frame.MessageClose) {
		log.Println("Error: connection wasn't closed propperly")
	}

	fmt.Println(msg)
	fmt.Println(time.Since(now))
}

func SendHelloMsg(conn *client.ClientConn) (*frame.Message, error) {
	err := conn.SendMessage(frame.MessageHello)
	if err != nil {
		return nil, err
	}

	msg, err := conn.ReceiveMessage()
	if err != nil {
		return nil, err
	}

	return msg, nil
}
