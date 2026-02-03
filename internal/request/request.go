package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/nicholasgswan/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type ParserState int

const (
	initialized ParserState = iota
	requestStateParsingHeaders
	parsingBody
	done
)
const crlf = "\r\n"
const bufferSize int = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{ParserState: initialized}
	req.Headers = headers.NewHeaders()
	b := make([]byte, bufferSize)
	readToIndex := 0

	for req.ParserState != done {
		if readToIndex >= len(b) {
			newBuf := make([]byte, len(b)*2)
			copy(newBuf, b)
			b = newBuf
		}
		br, err := reader.Read(b[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.ParserState == parsingBody {
					_, err := req.parse(b[readToIndex:])
					if err != nil {
						return nil, err
					}
					cl, err := strconv.Atoi(req.Headers.Get("content-length"))
					if err != nil {
						return nil, err
					}
					if cl != len(req.Body) {
						return nil, errors.New("body length does not match content length!")
					}
					break
				}
				if req.ParserState != done {
					return nil, fmt.Errorf("Incomplete Request, in state: %d, read n bytes on EOF: %d", req.ParserState, br)
				}
				break
			}
			return nil, err
		}
		readToIndex += br
		bp, err := req.parse(b[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(b, b[bp:])
		readToIndex -= bp

	}

	fmt.Println(req)
	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	rl, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return rl, idx + 2, nil
}

func requestLineFromString(req string) (*RequestLine, error) {
	rl := RequestLine{}
	arr := strings.Split(req, " ")

	if len(arr) != 3 {
		return &rl, errors.New("Invalid Request Line")
	}

	met := arr[0]
	alphanumeric := regexp.MustCompile("^[A-Z]*$")
	if !alphanumeric.MatchString(met) {
		return &rl, errors.New("Invalid Http method")
	}
	httpParts := strings.Split(arr[2], "/")
	var ver string
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return &rl, errors.New("Invalid Http version")
	} else {
		ver = httpParts[1]
	}
	rl.Method = met
	rl.RequestTarget = arr[1]

	rl.HttpVersion = ver
	return &rl, nil
}

func (r *Request) parse(data []byte) (int, error) {
	bytesRead := 0
	for r.ParserState != done {
		n, err := r.parseSingle(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += n
		if n == 0 {
			break
		}
	}

	return bytesRead, nil

}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case initialized:
		rl, bytesRead, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytesRead == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.ParserState = requestStateParsingHeaders
		return bytesRead, nil
	case requestStateParsingHeaders:

		n, isDone, err := r.Headers.Parse(data)

		if err != nil {
			return 0, err
		}
		if isDone {
			r.ParserState = parsingBody
		}
		return n, nil
	case parsingBody:
		cl := r.Headers.Get("content-length")
		if cl == "" || cl == "0" {
			r.ParserState = done
			return 0, nil
		}
		clNum, err := strconv.Atoi(cl)
		if err != nil {
			return 0, err
		}

		r.Body = append(r.Body, data...)
		if clNum != len(r.Body) {
			return len(data), nil
		}
		r.ParserState = done
		return len(r.Body), nil

	case done:
		return 0, errors.New("error: trying to read data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}

func (r *Request) String() string {
	str := "Request line:\n"
	str += fmt.Sprintf("- Method: %s\n", r.RequestLine.Method)
	str += fmt.Sprintf("- Target: %s\n", r.RequestLine.RequestTarget)
	str += fmt.Sprintf("- Version: %s\n", r.RequestLine.HttpVersion)
	str += "Headers:\n"
	for k, v := range r.Headers {
		str += fmt.Sprintf("- %s: %s\n", k, v)
	}
	return str

}
