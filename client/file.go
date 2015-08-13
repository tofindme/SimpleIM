package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {

	file, err := os.Open("./wstest.go")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "hello.txt")
	if err != nil {
		panic(err)
	}
	io.Copy(part, file)

	bodyType := writer.FormDataContentType()
	writer.Close()
	fmt.Printf("body type is %s\n", bodyType)
	fmt.Printf("body is %v", body)

	resp, err := http.Post("http://127.0.0.1:9000/upload", bodyType, body)
	if err != nil {
		panic(err)
	}
	fmt.Println("resp ", resp)

}
