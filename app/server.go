package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	server := "0.0.0.0"
	port := 4221

	l, err := net.Listen("tcp", fmt.Sprintf("%q:%q", server, port))
	if err != nil {
		fmt.Printf("Failed to bind to port %q\n", port)
		os.Exit(1)
	}

	defer l.Close()

	fmt.Printf("Server is listening on port %q\n", port)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %q", err.Error())
			os.Exit(1)
		}
		handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	msg := "HTTP/1.1 200 OK\r\n\r\n"
	no_bytes, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Error writing response: %q", err.Error())
	}
	fmt.Printf("Wrote %d bytes", no_bytes)
}
