package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/nicholasgswan/httpfromtcp/internal/request"
	"github.com/nicholasgswan/httpfromtcp/internal/response"
)

// const (
// 	response string = "HTTP/1.1 200 OK\r\n" +
// 		"Content-Type: text/plain\r\n" +
// 		//	"Content-Length: 13\r\n\r\n" +
// 		"\r\n" +
// 		"Hello World!\n"
// )

type Server struct {
	Open atomic.Bool
	port int
}

type HandlerError struct {
	statusCode response.StatusCode
	msg        string
}

func Serve(port int) (*Server, error) {
	s := &Server{Open: atomic.Bool{}, port: port}

	s.Open.Store(true)
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.Open.Store(false)
	if s.Open.Load() {
		return errors.New("Could not close server!")
	}
	return nil
}

func (s *Server) listen() {
	lis, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(s.port))
	if err != nil {
		log.Fatalf("Could not create Listener : %s", err.Error())
	}
	defer lis.Close()

	for s.Open.Load() {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Could not accept connection!")
			return
		}
		fmt.Println("Connection Open")

		go s.handle(conn)

	}

}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	fmt.Println("sending response")
	h := response.GetDefaultHeaders(0)
	response.WriteStatusLine(conn, response.OK)
	if err := response.WriteHeaders(conn, h); err != nil {
		fmt.Printf("error: %v/n", err)
	}
	return
}
