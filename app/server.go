package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	Method  string
	Target  string
	Version string

	Headers map[string]string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	server := "0.0.0.0"
	port := "4221"

	l, err := net.Listen("tcp", fmt.Sprintf("%v:%v", server, port))
	if err != nil {
		fmt.Printf("Failed to bind to port %q\n", port)
		os.Exit(1)
	}

	defer l.Close()

	fmt.Printf("Server is listening on port %q\n", port)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %q\n", err.Error())
			os.Exit(1)
		}
		handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	req := parseRequest(conn)

	msg := "HTTP/1.1 404 Not Found\r\n\r\n"

	if req.Target == "/"{
		msg = "HTTP/1.1 200 OK\r\n\r\n"
	}

	no_bytes, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Error writing response: %q\n", err.Error())
	}
	fmt.Printf("Wrote %d bytes\n", no_bytes)
}

func parseRequest(conn net.Conn) Request{

	readBuff := make([]byte, 1024)
	noOfBytes, err := conn.Read(readBuff)
	if err != nil {
		fmt.Printf("Error reading the request: %q\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Read %d bytes\n", noOfBytes)
	return parseRequestLine(readBuff)

}

func parseRequestLine(readBuff []byte) Request{
	requestStringParts := strings.Split(string(readBuff), "\r\n")
	req := Request{}
	for idx, part := range requestStringParts {
		if idx == 0 {
			fmt.Printf("Found request line %q\n", part)
			requestLineParts := strings.Split(part, " ")
			req.Method = requestLineParts[0]
			req.Target = requestLineParts[1]
			req.Version = requestLineParts[2]
		}
	}
	return req
}
