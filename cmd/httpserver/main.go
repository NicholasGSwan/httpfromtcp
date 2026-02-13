package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/nicholasgswan/httpfromtcp/internal/headers"
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
		headers.Replace("content-type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(internalServerErrorResponse))
	case "/video":
		videoRequest(w)
	default:
		if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/") {
			chunkedRequest(w, r)

		} else {
			w.WriteStatusLine(response.OK)
			headers := response.GetDefaultHeaders(len(successfulResponse))
			headers.Replace("content-type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody([]byte(successfulResponse))
		}

	}

}

func videoRequest(w response.Writer) {
	w.WriteStatusLine(response.OK)
	vid, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		fmt.Printf("Could not read file: %s", err.Error())
	}
	headers := response.GetDefaultHeaders(len(vid))
	//headers.Remove("Content-length")
	headers.Replace("content-type", "video/mp4")
	w.WriteHeaders(headers)
	w.WriteBody(vid)

}

func chunkedRequest(w response.Writer, r *request.Request) {
	w.WriteStatusLine(response.OK)
	h := response.GetDefaultChunkedHeaders()
	url := "https://httpbin.org/" + strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/")
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Could not get response : %s", err.Error())
	}
	defer resp.Body.Close()
	buf := make([]byte, 1024)
	body := make([]byte, 0)
	h.Set("Trailer", "X-Content-Length")
	h.Set("Trailer", "X-Content-SHA256")

	w.WriteHeaders(h)
	totalBytes := 0
	for {
		n, err := resp.Body.Read(buf)

		if n > 0 {
			var m int
			var err error

			m, err = w.WriteChunkedBody(buf[:n])

			if err != nil {
				fmt.Printf("Could not write chunk : %s", err.Error())
			}
			body = append(body, buf[:n]...)
			totalBytes += m
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("could not write chunk %s ", err.Error())
		}
	}
	n, err := w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Printf("Error finishing writing the body: %s", err.Error())
	}
	totalBytes += n
	hash := sha256.Sum256(body)
	t := headers.NewHeaders()
	t.Set("X-Content-Length", strconv.Itoa(len(body)))
	t.Set("X-Content-SHA256", fmt.Sprintf("%x", hash[:]))

	w.WriteTralers(t)

}
