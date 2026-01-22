package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)
		b := make([]byte, 8)
		//var offset int64 = 0
		var s strings.Builder
		var err error
		for err == nil {
			//var l int
			_, err = f.Read(b)
			arr := strings.Split(string(b), "\n")
			s.WriteString(arr[0])
			if len(arr) > 1 {
				out <- s.String()
				s.Reset()
				s.WriteString(arr[1])
			}
			if err == io.EOF {
				out <- s.String()
			}

			clear(b)
			//offset += 8
		}

	}()

	return out
}

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("oops")
		return
	}
	lines := getLinesChannel(f)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}

}
