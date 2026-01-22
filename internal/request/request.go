package request

import (
	"errors"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := Request{}
	b := make([]byte, 8)

	var s strings.Builder
	var err error
	for err == nil {
		_, err = reader.Read(b)

		arr := strings.Split(string(b), "\r\n")
		s.WriteString(arr[0])
		if len(arr) > 1 {
			break
		}

	}

	rl, err := parseRequestLine(s.String())

	req.RequestLine = *rl

	return &req, err
}

func parseRequestLine(req string) (*RequestLine, error) {
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
	ver := strings.Split(arr[2], "/")[1]
	if ver != "1.1" {
		return &rl, errors.New("Invalid Http version")
	}
	rl.Method = arr[0]
	rl.RequestTarget = arr[1]

	rl.HttpVersion = ver

	return &rl, nil
}
