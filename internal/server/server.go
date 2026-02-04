package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

const (
	response string = "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 13\r\n\r\n" +

		"Hello World!\n"
)

type Server struct {
	Open atomic.Bool
	port int
}

func Serve(port int) (*Server, error) {
	s := Server{Open: atomic.Bool{}, port: port}
	s.Open.Store(true)
	s.listen()
	return &s, nil
}

func (s *Server) Close() error {
	closed := s.Open.Swap(false)
	if !closed {
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
		//defer conn.Close()
		go func() {

			s.handle(conn)
		}()
	}

}

func (s *Server) handle(conn net.Conn) {
	fmt.Println("sending response")
	conn.Write([]byte(response))
}
