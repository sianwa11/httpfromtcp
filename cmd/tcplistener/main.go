package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error accepting connection: %s", err)
		}
		fmt.Println("connection has been accepted")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalln("error reading request:", err)
		}

		fmt.Printf("Request line: \n - Method: %s\n - Target: %s\n - Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)

		fmt.Println("connection has been closed")
	}

}
