package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

// const filePath = "messages.txt"

func getLinesChannel(conn net.Conn) <-chan string {
	lines := make(chan string)

	go func() {
		defer conn.Close()
		defer close(lines)

		currentLineContents := ""
		for {
			b := make([]byte, 8)
			n, err := conn.Read(b)
			if err != nil {
				if currentLineContents != "" {
					lines <- currentLineContents
				}

				if errors.Is(err, io.EOF) {
					break
				}

				fmt.Printf("an error occured: %s\n", err)
				return
			}

			str := string(b[:n])
			parts := strings.Split(str, "\n")

			for i := 0; i < len(parts)-1; i++ {
				lines <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
				currentLineContents = ""
			}

			currentLineContents += parts[len(parts)-1]
		}
	}()

	return lines
}

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

		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Println(line)
		}
		fmt.Println("connection has been closed")
	}

}
