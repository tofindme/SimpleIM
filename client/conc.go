package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"sync"
)

var (
	n    int
	room string
	user string
)

func init() {
	flag.IntVar(&n, "n", 100, "connect count")
	flag.StringVar(&user, "user", "test", "username to identify")
	flag.StringVar(&room, "room", "room1", "which room the users in")

}

func main() {

	var wait sync.WaitGroup
	flag.Parse()

	for i := 0; i < n; i++ {
		url := fmt.Sprintf("ws://localhost:9000/ws?user=%s%d&room=%s", user, i+1, room)
		wsConfig, err := websocket.NewConfig(url, "http://localhost:9000")
		if err != nil {
			fmt.Println("error : ", err.Error())
			return
		}
		wait.Add(1)
		fmt.Println("begin to conn")
		go func(i int) {
			defer wait.Done()
			for {
				wsConn, err := websocket.DialConfig(wsConfig)
				if err != nil {
					fmt.Println("conn error: ", err.Error())
					return
				}

				var message string
				if err := websocket.Message.Receive(wsConn, &message); err != nil {
					if err == io.EOF {
						fmt.Println("conn broken,bye!")
						break
					}
					fmt.Println("receive error : ", err.Error())
					continue
				}
				fmt.Printf(" received: %s\n", message)
			}
		}(i)
	}

	wait.Wait()

}
