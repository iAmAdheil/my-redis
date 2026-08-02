package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var vars = make(map[string]string)

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
			fmt.Printf("Error reading from connection: %s\n", err.Error())
			break
		}

		parts := RESPDecoder(n, in)
		// for _, v := range parts {
		// 	fmt.Printf("%q\n", v)
		// }

		switch parts[0] {
		case "*1":
			out = RESPEncoder("PONG", true)
			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "*2":
			if parts[1] == "$4" && strings.ToLower(parts[2]) == "echo" {
				out = RESPEncoder(parts[4], false)

				if _, err := conn.Write(out); err != nil {
					fmt.Printf("Error writing into connection: %s\n", err.Error())
				}
			} else if parts[1] == "$3" && strings.ToLower(parts[2]) == "get" {
				key := parts[4]

				val, ok := vars[key]
				if !ok {
					out = RESPEncoder("", false)
				} else {
					out = RESPEncoder(val, false)
				}

				if _, err := conn.Write(out); err != nil {
					fmt.Printf("Error writing into connection: %s\n", err.Error())
				}
			}

		case "*3":
			if parts[1] == "$3" && strings.ToLower(parts[2]) == "set" {
				key := parts[4]
				val := parts[6]

				vars[key] = val

				out = RESPEncoder("OK", true)

				if _, err := conn.Write(out); err != nil {
					fmt.Printf("Error writing into connection: %s\n", err.Error())
				}
			}
		}
	}
}

func RESPDecoder(n int, in []byte) []string {
	com := string(in[:n])

	// print incoming bytes -> testing
	// fmt.Printf("%q\n", in[:n])

	return strings.Split(com, "\r\n")
}

// simple string -> +{string}\r\n
// bulk string -> {string_len}\r\n{string}\r\n
func RESPEncoder(res string, simple bool) []byte {
	var s string
	if simple {
		s = fmt.Sprintf("+%s\r\n", res)
	} else {
		if len(res) == 0 {
			s = "$-1\r\n"
		} else {
			s = fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(res)), res)
		}
	}
	return []byte(s)
}
