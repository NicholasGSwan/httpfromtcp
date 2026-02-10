package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	handler  Handler
	listener net.Listener
	Open     atomic.Bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Msg        string
}

func (he HandlerError) Write(w io.Writer) {
	headers := response.GetDefaultHeaders(len(he.Msg))
	response.WriteStatusLine(w, he.StatusCode)
	if err := response.WriteHeaders(w, headers); err != nil {
		fmt.Printf("error: %v/n", err)
	}
	w.Write([]byte(he.Msg + "/r/n"))
}

func Serve(port int, handler Handler) (*Server, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	s := &Server{Open: atomic.Bool{}, listener: lis, handler: handler}

	s.Open.Store(true)
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.Open.Store(false)
	if s.Open.Load() {
		return errors.New("Could not close server!")
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {

	for s.Open.Load() {
		conn, err := s.listener.Accept()
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
	req, err := request.RequestFromReader(conn)
	if err != nil {
		he := &HandlerError{StatusCode: response.BadRequest, Msg: err.Error()}
		he.Write(conn)
		return
	}
	buf := &bytes.Buffer{}
	he := s.handler(buf, req)
	if he != nil {
		he.Write(conn)
		return
	}
	fmt.Println("sending response")
	h := response.GetDefaultHeaders(buf.Len())
	response.WriteStatusLine(conn, response.OK)
	if err := response.WriteHeaders(conn, h); err != nil {
		fmt.Printf("error: %v/n", err)
	}
	response.WriteBody(conn, buf)
	return
}
