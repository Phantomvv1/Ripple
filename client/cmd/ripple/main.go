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

	m, err := conn.SendMessage(msg)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(m)

	m, err = conn.SendMessage(msg) // 2 times in order to check cache speed
	log.Println("Second hello")
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(m)
	fmt.Println("\033[31mCache time: \033[m", time.Since(cacheCheck))

	log.Println("Sending close msg")
	m, err = conn.SendMessage(frame.MessageClose.Clone())
	log.Println("Sent close msg")
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
