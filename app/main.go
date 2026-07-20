package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	// in := make([]byte, 1024)

	// n, err := conn.Read(in)
	// if err != nil {
	// 	fmt.Println("Error reading data from connection: ", err.Error())
	// }

	// fmt.Println("Length of incoming data:", n)
	// // fmt.Println("Incoming data:", in)

	// in = bytes.TrimSpace(in)

	res := []byte("+PONG\r\n")
	_, err = conn.Write(res)
	if err != nil {
		fmt.Println("Error writing data to connection: ", err.Error())
	}

	err = conn.Close()
	if err != nil {
		fmt.Println("Error while closing connnection", err.Error())
	}
}
