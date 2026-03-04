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

	msg, err := frame.NewMessage([]byte("Hello"), frame.RequestMsg, frame.CachableFlag|frame.AuthEnabledFlag, 0)
	if err != nil {
		log.Println(err)
		return
	}

	conn, err := client.Dial(":42069")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	cacheCheck := time.Now()

	m, err := SendHelloMsg(conn, msg)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(m)

	m, err = SendHelloMsg(conn, msg) // 2 times in order to check cache speed
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(m)
	fmt.Println("\033[31mCache time: \033[m", time.Since(cacheCheck))

	err = conn.SendMessage(frame.MessageClose)
	if err != nil {
		log.Println(err)
		return
	}

	m, err = conn.ReceiveMessage()
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(m)
	fmt.Println(time.Since(now))

	if !m.Equals(*frame.MessageClose) {
		log.Println("Error: connection wasn't closed propperly")
	}

}

func SendHelloMsg(conn *client.ClientConn, message *frame.Message) (*frame.Message, error) {
	err := conn.SendMessage(message)
	if err != nil {
		return nil, err
	}

	msg, err := conn.ReceiveMessage()
	if err != nil {
		return nil, err
	}

	return msg, nil
}
