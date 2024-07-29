package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
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

type Response struct {
	Version       string
	StatusCode    string
	StatusMessage string
	Headers       map[string]string
	Body          string
}

func (r *Response)ToMessage() string {
	var sb strings.Builder
	for key, value := range r.Headers {
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\r\n")
	}
	return fmt.Sprintf("%s %s %s\r\n%s\r\n%s", r.Version, r.StatusCode, r.StatusMessage, sb.String(), r.Body)
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

	resp := Response{
		Version: "HTTP/1.1",
		StatusCode: "404",
		StatusMessage: "Not Found",
		Headers: make(map[string]string),
	}

	supported := isEncryptionHeaderSupported(req.Headers)

	if supported {
		resp.Headers["Content-Encoding"] = "gzip"
	}

	if req.Target == "/" {
		resp.StatusCode = "200"
		resp.StatusMessage = "OK"
	}

	if strings.HasPrefix(req.Target, "/echo/") {
		toEcho := strings.Split(req.Target, "/echo/")
		resp.StatusCode = "200"
		resp.StatusMessage = "OK"
		resp.Headers["Content-Type"] = "text/plain"
		resp.Headers["Content-Length"] = strconv.Itoa(len(toEcho[1]))
		resp.Body = toEcho[1]
	}

	if req.Target == "/user-agent" {
		ua := req.Headers["User-Agent"]
		resp.StatusCode = "200"
		resp.StatusMessage = "OK"
		resp.Headers["Content-Type"] = "text/plain"
		resp.Headers["Content-Length"] = strconv.Itoa(len(ua))
		resp.Body = ua
	}

	if strings.HasPrefix(req.Target, "/files/") {
		resp = processFilesRequest(req, strings.Split(req.Target, "/files/")[1], filesDir)
	}

	msg := resp.ToMessage()

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
				req.Body = req.Body + strings.Trim(part, "\x00")
			}
		}
	}
	return req
}

func processFilesRequest(req Request, fileName string, filesDir string) Response {
	absoluteFilePath := fmt.Sprintf("%v/%v", filesDir, fileName)

	if req.Method == "GET" {
		return getFile(absoluteFilePath)
	}
	if req.Method == "POST" {
		return createFile(absoluteFilePath, req)
	}
	resp := Response{
		Version: "HTTP/1.1",
		StatusCode: "404",
		StatusMessage: "Not Found",
		Headers: make(map[string]string),
	}
	return resp
}

func getFile(absoluteFilePath string) Response {
	data, err := os.ReadFile(absoluteFilePath)
	resp := Response{
		Version: "HTTP/1.1",
		StatusCode: "404",
		StatusMessage: "Not Found",
		Headers: make(map[string]string),
	}
	if err != nil {
		fmt.Printf("error %q reading file %v", err.Error(), absoluteFilePath)
		return resp
	}
	resp.StatusCode = "200"
	resp.StatusMessage = "OK"
	resp.Headers["Content-Type"] = "application/octet-stream"
	resp.Headers["Content-Length"] = strconv.Itoa(len(string(data)))
	resp.Body = string(data)
	return resp
}

func createFile(absoluteFilePath string, req Request) Response {
	err := os.WriteFile(absoluteFilePath, []byte(req.Body), 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %q\n", err.Error())
		os.Exit(1)
	}
	resp := Response {
		Version: "HTTP/1.1",
		StatusCode: "201",
		StatusMessage: "Created",
	}
	return resp
}

func isEncryptionHeaderSupported(headers map[string]string) bool {
	for key, value := range headers {
		if key == "Accept-Encoding" {
			return strings.Contains(value, "gzip")
		}
	}
	return false
}
