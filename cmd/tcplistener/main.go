package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nicholasgswan/httpfromtcp/internal/request"
)

// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	out := make(chan string, 1)

// 	go func() {
// 		defer f.Close()
// 		defer close(out)
// 		b := make([]byte, 8)
// 		//var offset int64 = 0
// 		var s strings.Builder
// 		var err error
// 		for err == nil {
// 			//var l int
// 			_, err = f.Read(b)
// 			arr := strings.Split(string(b), "\n")
// 			s.WriteString(arr[0])
// 			if len(arr) > 1 {
// 				out <- s.String()
// 				s.Reset()
// 				s.WriteString(arr[1])
// 			}
// 			if err == io.EOF {
// 				out <- s.String()
// 			}

// 			clear(b)
// 			//offset += 8
// 		}

// 	}()

// 	return out
// }

func main() {
	//f, err := os.Open("messages.txt")

	lis, err := net.Listen("tcp", "127.0.0.1:42069")

	if err != nil {
		log.Fatalf("oops... %s\n", err.Error())
	}
	defer lis.Close()
	for true {
		conn, err1 := lis.Accept()
		if err1 != nil {
			fmt.Println("oops")
			return
		}
		fmt.Println("Connection has been accepted!")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("invalid request: %s\n", err.Error())
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)

		fmt.Println("Connection is closed")
	}

}
