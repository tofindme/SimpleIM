package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	// "net"
	// "io"
	"sync"
	"time"
)

//============================================
//user struct
//

type YbUser struct {
	cntx   *YbRoom
	Name   string
	Conn   *websocket.Conn
	send   chan []byte
	bclose chan bool
}

func NewUser(conn *websocket.Conn, name string) *YbUser {
	user := new(YbUser)
	user.Conn = conn
	user.Name = name
	user.send = make(chan []byte)
	user.bclose = make(chan bool)
	return user
}

func (u *YbUser) GetName() string {
	return u.Name
}

func (u *YbUser) AttachRoom(room *YbRoom) {
	u.cntx = room
}

func (u *YbUser) DetachRoom() {
	if u.cntx != nil {
		u.cntx.Unregister <- u
	}
}

func (u *YbUser) Start() {
	// defer func() {
	// 	fmt.Printf("websocket stopded for user[%s]\n", u.Name)
	// 	u.DetachRoom()
	// 	u.Conn.Close()
	// }()

	fmt.Printf("user[%s] started\n", u.Name)

	userTick := time.NewTicker(time.Second * 3)
	go func() {
		for {
			select {
			case msg := <-u.send:
				err := u.Conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					fmt.Printf("send msg failed for user[%s]\n", u.Name)
					return
				}
			case <-u.bclose:
				fmt.Printf("conn exit for user[%s]\n", u.Name)
				return
			case <-userTick.C:
				err := u.Conn.WriteMessage(websocket.TextMessage, []byte("heartbeat"))
				if err != nil {
					fmt.Printf("heartbeat failed for user[%s]\n", u.Name)
					u.Stop()
					return
				}
				fmt.Printf("user's tick for user[%s]\n", u.Name)
			}
		}
	}()

	for {
		_, message, err := u.Conn.ReadMessage()
		if err != nil {
			fmt.Printf("broken conn user[%s]\n", u.Name)
			u.Stop()
			break
		}
		if string(message) == "close" {
			u.Stop()
			break
		}

		fmt.Printf("user[%s] recive msg : %s\n", u.Name, string(message))
		u.cntx.BroadCast(message)
	}

}

func (u *YbUser) Stop() {

	// check if the ChanDone is closed
	isClosed := false
	select {
	case _, ok := <-u.bclose:
		if !ok {
			isClosed = true
		}
		fmt.Println("channel recive")
	default:
		fmt.Printf("channel not value\n")
	}

	if isClosed {
		return
	}

	u.DetachRoom()
	u.Conn.Close()
	close(u.bclose)
}

//============================================
//room struct
//

var GRooms = make(map[string]*YbRoom)

//a room to chat
type YbRoom struct {
	sync.Mutex                    //lock to hold
	name       string             //romm name
	users      map[string]*YbUser //user map name-user
	Msg        chan []byte        //room msg
	Register   chan *YbUser       //new user
	Unregister chan *YbUser       //use leaft
}

func NewRoom(name string) *YbRoom {
	room := new(YbRoom)
	room.SetName(name)
	room.users = make(map[string]*YbUser)
	room.Msg = make(chan []byte, 10)
	room.Register = make(chan *YbUser)
	room.Unregister = make(chan *YbUser)
	return room
}

//get room
//reciver should be pointer,otherwore the r.name was empty.  why?
func (r *YbRoom) GetName() string {
	return r.name
}

//set name
func (r *YbRoom) SetName(n string) {
	r.name = n
}

func (r *YbRoom) AddUser(user *YbUser) {
	r.Lock()
	defer r.Unlock()
	//exist?
	if _, ok := r.users[user.GetName()]; !ok {
		r.users[user.GetName()] = user
	}
}

func (r *YbRoom) RemoveUser(user *YbUser) {
	r.Lock()
	defer r.Unlock()
	//
	if _, ok := r.users[user.GetName()]; ok {
		delete(r.users, user.GetName())
	}

}

func (r *YbRoom) BroadCast(msg []byte) {
	r.Lock()
	defer r.Unlock()

	//send data to the users
	for _, user := range r.users {
		//send to send channel
		user.send <- msg
	}

}

func (r *YbRoom) Run() {
	timer := time.NewTicker(time.Second * 3)
	//event waiting
	for {
		select {
		case user := <-r.Register:
			fmt.Printf("new users comes user[%s] in room[%s]\n", user.GetName(), r.GetName())
			r.AddUser(user)
		case user := <-r.Unregister:
			fmt.Printf("user say goodbye user[%s]\n", user.GetName())
			r.RemoveUser(user)
		case msg := <-r.Msg:
			r.BroadCast(msg)
		case <-timer.C:
			//r.Lock()
			if len(r.users) == 0 {
				fmt.Printf("emppty room[%s],should deadth!\n", r.GetName())
				r.Destory()
				delete(GRooms, r.GetName())
				return
			}

			fmt.Printf("room's tick in this room[%s]\n", r.GetName())
		}
	}
}

func (r *YbRoom) Destory() {
	//close user conn
	for _, user := range r.users {
		//user.close
		user.send <- []byte("close")
	}
}
