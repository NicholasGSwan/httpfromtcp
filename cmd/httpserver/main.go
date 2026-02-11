package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		w.WriteStatusLine(response.BadRequest)
		headers := response.GetDefaultHeaders(len(badRequestResponse))
		headers.Set("content-type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(badRequestResponse))
	case "/myproblem":
		w.WriteStatusLine(response.ServerError)
		headers := response.GetDefaultHeaders(len(internalServerErrorResponse))
		headers.Set("content-type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(internalServerErrorResponse))
	default:
		if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/") {
			chunkedRequest(w, r)

		} else {
			w.WriteStatusLine(response.OK)
			headers := response.GetDefaultHeaders(len(successfulResponse))
			headers.Set("content-type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody([]byte(successfulResponse))
		}

	}

}

func chunkedRequest(w response.Writer, r *request.Request) {
	var toBreak bool
	w.WriteStatusLine(response.OK)
	h := response.GetDefaultChunkedHeaders()
	w.WriteHeaders(h)
	chunk, err := strconv.Atoi(strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/stream/"))
	if err != nil {
		fmt.Printf("Could not convert string : %s", err.Error())
	}
	url := fmt.Sprintf("https://httpbin.org/stream/%d", chunk)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Could not get response : %s", err.Error())
	}
	defer resp.Body.Close()
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				toBreak = true
			} else {
				fmt.Printf("Could not read response : %s", err.Error())
			}
		}
		written := 0
		for written < n {
			var m int
			var err error
			if written+chunk > n {
				m, err = w.WriteChunkedBody(buf[written:n])
			} else {
				m, err = w.WriteChunkedBody(buf[written : written+chunk])
			}

			if err != nil {
				fmt.Printf("Could not write chunk : %s", err.Error())
			}
			written += m
		}
		if toBreak {
			break
		}
	}

	w.WriteChunkedBodyDone()
}
