package goscgi

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"
	"time"
)

// not a real test :|
func Test_All(t *testing.T) {
	log.Println("test starting..")

	srv := NewServer(NewSettings())
	srv.AddHandler("/cgi/", requestHandler)

	go startServer(srv)
	time.Sleep(time.Second)

	startClient()
	close(srv.Close) // signal server to stop listening

	log.Println("test done")
}

func startServer(srv *Server) {
	err := srv.ListenTcp(":8080")
	if err != nil {
		log.Println("startServer, srv.ListenTcp", err.Error())
	}
}

func requestHandler(req *Request) *Response {
	log.Println()
	log.Println("handling request:")
	for k, v := range req.Header {
		log.Println(k, " = ", v)
	}
	if len(req.Content) > 0 {
		log.Println(string(req.Content))
	}

	return NewResponse(RespCodeOK, RespTypeHtml, []byte("this is a test response"))
}

func startClient() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Println("startClient, net.Dial:", err.Error())
		return
	}
	defer conn.Close()

	header := make(map[string]string)
	header["CONTENT_LENGTH"] = "" // will be computed later
	header["REQUEST_URI"] = "/cgi/test.cgi"
	header["REQUEST_METHOD"] = "GET"
	header["CONTENT_TYPE"] = "text/plain"
	content := "this is a test request"

	sendRequest(conn, header, content)
	time.Sleep(time.Second)

	log.Println()
	log.Println("server response:")

	var buff [80]byte
	for {
		readCnt, _ := conn.Read(buff[:])
		if readCnt == 0 {
			return
		}
		log.Println(string(buff[:readCnt]))
	}
}

// http://www.python.ca/scgi/protocol.txt
func sendRequest(conn net.Conn, header map[string]string, content string) {
	if len(content) > 0 {
		header["CONTENT_LENGTH"] = strconv.Itoa(len(content))
	}

	headerSize := 0
	for k, v := range header {
		headerSize += len(k) + 1 // include the separator byte
		headerSize += len(v) + 1
	}
	headerSizeStr := strconv.Itoa(headerSize)

	log.Println("sendRequest:")
	conn.Write([]byte(headerSizeStr))
	log.Println("sent headerSize:", headerSizeStr)
	fmt.Fprint(conn, ":")

	zero := []byte{0}
	for k, v := range header {
		fmt.Fprint(conn, k)
		conn.Write(zero)
		fmt.Fprint(conn, v)
		conn.Write(zero)
		log.Println("sent ", k, "=", v)
	}

	fmt.Fprint(conn, ",")
	fmt.Fprint(conn, content)
	log.Println("sent content:", content)
}
