package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var vars = make(map[string]string)
var lists = make(map[string]*[]string)

func UpdateList(listkey string, val string) int {
	var count int

	l, ok := lists[listkey]
	if !ok {
		nl := &[]string{val} // new list
		lists[listkey] = nl
		count = 1
	} else {
		*l = append(*l, val)
		count = len(*l)
	}

	return count
}

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
			out = RESPEncoder("PONG", 0)
			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "*2":
			if parts[1] == "$4" && strings.ToLower(parts[2]) == "echo" {
				out = RESPEncoder(parts[4], 1)

				if _, err := conn.Write(out); err != nil {
					fmt.Printf("Error writing into connection: %s\n", err.Error())
				}
			} else if parts[1] == "$3" && strings.ToLower(parts[2]) == "get" {
				key := parts[4]

				val, ok := vars[key]
				if !ok {
					out = RESPEncoder("", 1)
				} else {
					out = RESPEncoder(val, 1)
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

				out = RESPEncoder("OK", 0)

				if _, err := conn.Write(out); err != nil {
					fmt.Printf("Error writing into connection: %s\n", err.Error())
				}
			} else if parts[1] == "$5" && strings.ToLower(parts[2]) == "rpush" {
				listkey := parts[4]
				val := parts[6]

				count := UpdateList(listkey, val)

				out = RESPEncoder(strconv.Itoa(count), 2)

				if _, err := conn.Write(out); err != nil {
					fmt.Printf("Error writing into connection: %s\n", err.Error())
				}
			}

		case "*5":
			if parts[1] == "$3" && strings.ToLower(parts[2]) == "set" {
				key := parts[4]
				val := parts[6]

				// ex -> second
				// px -> millisecond
				t := parts[8]
				d, err := strconv.ParseInt(parts[10], 10, 64)
				if err != nil {
					fmt.Printf("Error converting duration to int: %s\n", err.Error())
				}

				var m time.Duration
				if strings.ToLower(t) == "px" {
					m = time.Millisecond
				} else {
					m = time.Second
				}

				vars[key] = val
				go Expire(time.Duration(d)*m, key)

				out = RESPEncoder("OK", 0)

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

// simple string -> +{string}\r\n -> 0
// bulk string -> {string_len}\r\n{string}\r\n -> 1
// RESP integer -> :{integer (sent as a string)}\r\n -> 2
func RESPEncoder(res string, t int) []byte {
	var s string
	if t == 0 {
		s = fmt.Sprintf("+%s\r\n", res)
	} else if t == 2 {
		s = fmt.Sprintf(":%s\r\n", res)
	} else if t == 1 {
		if len(res) == 0 {
			s = "$-1\r\n"
		} else {
			s = fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(res)), res)
		}
	}
	return []byte(s)
}

func Expire(t time.Duration, key string) {
	time.Sleep(t)
	delete(vars, key)
}
