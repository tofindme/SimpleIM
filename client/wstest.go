package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	// "os"
)

var (
	userName string
	room     string
)

func init() {

	flag.StringVar(&userName, "username", "test", "username to identify")
	flag.StringVar(&room, "room", "room1", "which room the users in")

}

func main() {

	flag.Parse()

	url := fmt.Sprintf("ws://localhost:9000/ws?user=%s&room=%s", userName, room)
	wsConfig, err := websocket.NewConfig(url, "http://localhost:9000")
	if err != nil {
		fmt.Println("error : ", err.Error())
		return
	}

	wsConn, err := websocket.DialConfig(wsConfig)
	if err != nil {
		fmt.Println("conn error: ", err.Error())
		return
	}

	read := make(chan int)
	//for read
	go func() {
		for {
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
			read <- 1
		}
	}()

	for i := 0; i < 10; i++ {
		msg := userName + " : " + "i am comming "
		msg = fmt.Sprintf("%s %d", msg, i)
		err := websocket.Message.Send(wsConn, msg)
		if err != nil {
			fmt.Println("send failed!")
		}
		fmt.Println("send success,wait read")
		<-read
	}

}
