package request

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type ParserState int

const (
	initialized ParserState = iota
	done
)
const crlf = "\r\n"
const bufferSize int = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{ParserState: initialized}
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
				req.ParserState = done
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
	var err error
	switch r.ParserState {
	case initialized:
		var rl *RequestLine
		rl, bytesRead, err = parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytesRead == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.ParserState = done
	case done:
		return 0, errors.New("error: trying to read data in a done state.")
	default:
		return 0, errors.New("unknown state")
	}

	return bytesRead, err

}
