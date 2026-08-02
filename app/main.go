package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go HandleConn(conn)
	}
}

func HandleConn(conn net.Conn) {
	for {
		in := make([]byte, 100)
		var out []byte

		n, err := conn.Read(in)
		if err != nil {
			panic(fmt.Sprintf("Error reading from connection: %s\n", err.Error()))
		}

		parts := RESPDecoder(n, in)
		for _, v := range parts {
			fmt.Printf("%q\n", v)
		}

		switch parts[0] {
		case "*1":
			out = []byte("+PONG\r\n")
			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "*2":
			if parts[1] == "$4" && strings.ToLower(parts[2]) == "echo" {
				outs := fmt.Sprintf("%s\r\n%s\r\n", parts[3], parts[4])
				out = []byte(outs)

				if _, err := conn.Write(out); err != nil {
					fmt.Printf("Error writing into connection: %s\n", err.Error())
				}
			}
		}
	}
}

func RESPDecoder(n int, in []byte) []string {
	// n-1 to remove the trailing \n
	com := string(in[:n])
	return strings.Split(com, "\r\n")
}
