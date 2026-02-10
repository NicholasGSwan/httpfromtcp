package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nicholasgswan/httpfromtcp/internal/request"
	"github.com/nicholasgswan/httpfromtcp/internal/response"
	"github.com/nicholasgswan/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handleReq)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handleReq(w io.Writer, r *request.Request) *server.HandlerError {
	if r.RequestLine.RequestTarget == "/yourproblem" {
		he := &server.HandlerError{Msg: "Your problem is not my problem", StatusCode: response.BadRequest}
		he.Write(w)
		return he
	} else if r.RequestLine.RequestTarget == "/myproblem" {
		he := &server.HandlerError{Msg: "Woopsie, my bad", StatusCode: response.ServerError}
		he.Write(w)
		return he
	} else {
		w.Write([]byte("All good, frfr\r\n"))
		return nil
	}

}
