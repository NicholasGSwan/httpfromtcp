package response

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/nicholasgswan/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK          StatusCode = 200
	BadRequest  StatusCode = 400
	ServerError StatusCode = 500
)

type writerStatus int

const (
	writeStatusLine writerStatus = iota
	writeHeaders
	writeBody
	writeTrailers
	writeComplete
)

type Writer struct {
	writer io.Writer
	status writerStatus
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error
	switch statusCode {
	case OK:
		_, err = w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case BadRequest:
		_, err = w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case ServerError:
		_, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		fmt.Println("Did not write statuscode")
	}
	return err
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var err error
	if w.status != writeStatusLine {
		return errors.New("Writer not in Status Line write state")
	}
	switch statusCode {
	case OK:
		_, err = w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		w.status = writeHeaders
	case BadRequest:
		_, err = w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		w.status = writeHeaders
	case ServerError:
		_, err = w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		w.status = writeHeaders
	default:
		fmt.Println("Did not write statuscode")
	}
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.Headers{}
	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")
	return h

}

func GetDefaultChunkedHeaders() headers.Headers {
	h := headers.Headers{}
	h.Set("Transfer-Encoding", "chunked")
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {

		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.status != writeHeaders {
		return errors.New("Writer not in write headers state")
	}

	return w.headerHelperFunc(headers)
}

func WriteBody(w io.Writer, body *bytes.Buffer) error {
	_, err := w.Write(body.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.status != writeBody {
		return 0, errors.New("Writer not in write body state")
	}
	n, err := w.writer.Write(body)
	if err != nil {
		return 0, err
	}
	w.status = writeTrailers
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.status != writeBody {
		return 0, errors.New("Writer not in write body state")
	}
	n, err := w.writer.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	if err != nil {
		return 0, err
	}
	m, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	k, err := w.writer.Write([]byte("\r\n"))
	return n + m + k, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	end := []byte("0\r\n")
	n, err := w.writer.Write(end)
	if err != nil {
		return 0, err
	}
	w.status = writeTrailers
	return n, nil
}

func (w *Writer) WriteTralers(headers headers.Headers) error {
	if w.status != writeTrailers {
		return errors.New("Writer not in write trailer state")
	}
	return w.headerHelperFunc(headers)
}

func (w *Writer) headerHelperFunc(headers headers.Headers) error {
	for k, v := range headers {
		fmt.Printf("%s: %s\n", k, v)
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	switch w.status {
	case writeHeaders:
		w.status = writeBody
	case writeTrailers:
		w.status = writeComplete
	}
	return nil
}

func NewWriter(w io.Writer) Writer {
	return Writer{writer: w, status: writeStatusLine}
}
