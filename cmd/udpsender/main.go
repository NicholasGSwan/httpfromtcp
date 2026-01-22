package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("udpsender running")
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")

	if err != nil {
		fmt.Println("oop")
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		fmt.Println("oop, that bad")
		fmt.Println(err.Error())
		return
	}
	defer conn.Close()

	r := bufio.NewReader(os.Stdin)

	for true {
		fmt.Print(">")
		s, err := r.ReadString('\n')
		printError(err)
		_, err = conn.Write([]byte(s))
		printError(err)

	}
}

func printError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
