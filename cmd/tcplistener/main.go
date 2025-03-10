package main

import (
	"fmt"
	"log"
	"net"

	"github.com/mrcordova/httpfromtcp/internal/request"
)



const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		requestLine, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("failed to get requestline")
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", requestLine.RequestLine.Method)
		fmt.Printf("- Target: %s\n", requestLine.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", requestLine.RequestLine.HttpVersion)
	
	}
}

