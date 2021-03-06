package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type RouterHandler struct{}

var chat = template.Must(template.ParseFiles("chat.html"))

func (r RouterHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	//dispatch action
	switch req.URL.Path {
	case "/login", "/":
		r.ProcessLogin(w, req)
	case "/user":
		r.ProcessUser(w, req)
	case "/upload":
		r.ProcessUpload(w, req)
	case "/ws":
		r.ProcessWebsocket(w, req)
	// case "/download":
	// http.ServeFile(w, req, "./uploads")
	// http.FileServer(http.Dir("./")).ServeHTTP()
	default:
		http.Error(w, "not found for the url", 500)
	}
}

func (r RouterHandler) ProcessUpload(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("req.Method is %s\n", req.Method)
	if req.Method == "POST" {
		req.ParseMultipartForm(32 << 20)
		file, handler, err := req.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./uploads/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func (r RouterHandler) ProcessLogin(w http.ResponseWriter, req *http.Request) {
	//cookie format is k=v;k=v;k=v

	// Login.Execute(w, nil)
	f, _ := os.Open("login.html")

	stream, _ := ioutil.ReadAll(f)

	fmt.Fprintln(w, string(stream))

}

func (r RouterHandler) ProcessUser(w http.ResponseWriter, req *http.Request) {
	//cookie format is k=v;k=v;k=v
	// w.Header().Set("Cookie", "name=yibin;age=20")

	user := req.FormValue("user")
	if user == "" {
		http.Error(w, "who are you?", 500)
		return
	}

	room := req.FormValue("room")
	if room == "" {
		http.Error(w, "please chose the room", 500)
		return
	}
	type temp struct {
		User string
		Room string
		Host string
	}

	data := temp{User: user, Room: room, Host: "127.0.0.1:9000"}

	chat.Execute(w, data)

}

func (r RouterHandler) ProcessWebsocket(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	room := req.FormValue("room")
	if room == "" {
		http.Error(w, "please chose the room!", 500)
	}

	userName := req.FormValue("user")
	if userName == "" {
		http.Error(w, "who are you?", 500)
	}

	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, "create websocket error", 500)
	}
	user := NewUser(ws, userName)

	fmt.Println("remote add is ", ws.RemoteAddr())

	if hoom, ok := GRooms[room]; !ok {
		// fmt.Println("why no value?", room)
		Room := NewRoom(room)
		GRooms[room] = Room
		go Room.Run()
		// time.Sleep(time.Second * 1)
		Room.Register <- user
		user.AttachRoom(Room)
		go user.Start()
	} else {
		hoom.Register <- user
		user.AttachRoom(hoom)
		go user.Start()
	}

}

func main() {

	handler := RouterHandler{}

	// http.Handle("/download", http.FileServer(http.Dir("/tmp")))
	server := &http.Server{
		Addr:         ":9000",
		Handler:      handler,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	go func() {
		fmt.Println(http.ListenAndServe(":9001", http.FileServer(http.Dir("./uploads"))))
	}()

	// http.HandleFunc("/", ProcessLogin)
	// http.HandleFunc("/login", ProcessLogin)
	// http.HandleFunc("/user", ProcessUser)
	// http.HandleFunc("/upload", ProcessUpload)
	// http.HandleFunc("/ws", ProcessWebsocket)
	// http.Handle("/download", http.FileServer(http.Dir("./uploads")))

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("litern failed with error : %s\n", err.Error())
	}

}
