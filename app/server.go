package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	msg := "HTTP/1.1 200 OK\r\n\r\n"
	no_bytes, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
	fmt.Printf("Wrote %d bytes", no_bytes)
}
