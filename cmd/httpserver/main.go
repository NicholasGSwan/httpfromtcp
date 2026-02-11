package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nicholasgswan/httpfromtcp/internal/request"
	"github.com/nicholasgswan/httpfromtcp/internal/response"
	"github.com/nicholasgswan/httpfromtcp/internal/server"
)

const port = 42069

const badRequestResponse = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalServerErrorResponse = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const successfulResponse = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

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

func handleReq(w response.Writer, r *request.Request) {
	if r.RequestLine.RequestTarget == "/yourproblem" {
		w.WriteStatusLine(response.BadRequest)
		headers := response.GetDefaultHeaders(len(badRequestResponse))
		headers.Set("content-type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(badRequestResponse))

	} else if r.RequestLine.RequestTarget == "/myproblem" {
		w.WriteStatusLine(response.ServerError)
		headers := response.GetDefaultHeaders(len(internalServerErrorResponse))
		headers.Set("content-type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(internalServerErrorResponse))
	} else {
		w.WriteStatusLine(response.OK)
		headers := response.GetDefaultHeaders(len(successfulResponse))
		headers.Set("content-type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(successfulResponse))
	}

}
