package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	// Request line
	Method  string
	Target  string
	Version string
	// Headers
	Headers map[string]string
	// Body
	Body string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	server := "0.0.0.0"
	port := "4221"

	filesDir := flag.String("directory", "/tmp", "Directory where the files are stored")
	flag.Parse()

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
		go handleConnection(conn, *filesDir)
	}

}

func handleConnection(conn net.Conn, filesDir string) {

	defer conn.Close()

	req := parseRequest(conn)

	msg := "HTTP/1.1 404 Not Found\r\n\r\n"

	if req.Target == "/" {
		msg = "HTTP/1.1 200 OK\r\n\r\n"
	}

	if strings.HasPrefix(req.Target, "/echo/") {
		toEcho := strings.Split(req.Target, "/echo/")
		msg = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%v", len(toEcho[1]), toEcho[1])
	}

	if req.Target == "/user-agent" {
		ua := req.Headers["User-Agent"]
		msg = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%v", len(ua), ua)
	}

	if strings.HasPrefix(req.Target, "/files/") {
		msg = processFilesRequest(req, strings.Split(req.Target, "/files/")[1], filesDir)
	}

	no_bytes, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Error writing response: %q\n", err.Error())
	}
	fmt.Printf("Wrote %d bytes\n", no_bytes)
}

func parseRequest(conn net.Conn) Request {

	readBuff := make([]byte, 1024)
	noOfBytes, err := conn.Read(readBuff)
	if err != nil {
		fmt.Printf("Error reading the request: %q\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Read %d bytes\n", noOfBytes)
	return parseRequestLine(readBuff)

}

func parseRequestLine(readBuff []byte) Request {
	requestStringParts := strings.Split(string(readBuff), "\r\n")
	req := Request{}
	req.Headers = make(map[string]string)
	headerComplete := false
	for idx, part := range requestStringParts {
		if idx == 0 {
			fmt.Printf("Found request line %q\n", part)
			requestLineParts := strings.Split(part, " ")
			req.Method = requestLineParts[0]
			req.Target = requestLineParts[1]
			req.Version = requestLineParts[2]
		} else {
			// We are now dealing with headers or body
			if len(part) == 0 {
				headerComplete = true
			}
			// We are still processing headers
			if !headerComplete {
				headerParts := strings.Split(part, ":")
				req.Headers[headerParts[0]] = strings.Trim(headerParts[1], " ")
			} else {
				// This must be the body
				req.Body = req.Body + part
			}
		}
	}
	return req
}

func processFilesRequest(req Request, fileName string, filesDir string) string {
	absoluteFilePath := fmt.Sprintf("%v/%v", filesDir, fileName)

	if req.Method == "GET" {
		return getFile(absoluteFilePath)
	} 
	if (req.Method == "POST") {
		return createFile(absoluteFilePath, req)
	}
	return "HTTP/1.1 400 BAD REQUEST\r\n\r\n"
}

func getFile(absoluteFilePath string) string {
	data, err := os.ReadFile(absoluteFilePath)
	if err != nil {
		fmt.Printf("error %q reading file %v", err.Error(), absoluteFilePath)
		return "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(string(data)), string(data))
}

func createFile(absoluteFilePath string, req Request) string {
	err := os.WriteFile(absoluteFilePath, []byte(req.Body), 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %q\n", err.Error())
		os.Exit(1)
	}
	return "HTTP/1.1 201 Created\r\n\r\n"
}
