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

	for {
		in := make([]byte, 100)
		_, err := conn.Read(in)
		if err != nil {
			fmt.Printf("Error reading from connection: %s\n", err.Error())
		}

		// _ := strings.Split(strings.TrimSpace(string(in)), "\n")[0]
		out := []byte("+PONG\r\n")
		if _, err := conn.Write(out); err != nil {
			fmt.Printf("Error writing into connection: %s\n", err.Error())
		}
	}

	// err = conn.Close()
	// if err != nil {
	// 	fmt.Println("Error while closing connnection", err.Error())
	// }
}
