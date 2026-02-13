package headers

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	endInd := bytes.Index(data, []byte(crlf))
	if endInd == 0 {
		return 2, true, nil
	}
	if endInd == -1 {
		return 0, false, nil
	}
	str := strings.TrimSpace(string(data[:endInd]))
	arr := strings.Split(str, " ")

	if !strings.HasSuffix(arr[0], ":") {
		return 0, false, errors.New("Malformed Header-line")
	}
	key := arr[0][:len(arr[0])-1]
	r, err := regexp.Compile(`[a-zA-Z0-9!#$%&*+\-.^_|~\x60]+`)
	if match, _ := regexp.MatchString(r.String(), key); !match {
		return 0, false, errors.New("invalid header key")
	}

	h.Set(key, arr[1])

	return endInd + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)

	if h[key] != "" && key != "content-type" {
		h[key] = h[key] + ", " + value
	} else {
		h[key] = value
	}

}

func (h Headers) Get(key string) string {
	key = strings.ToLower(key)
	return h[key]
}

func (h Headers) Replace(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Remove(key string) {
	delete(h, key)
}
