package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var vars = make(map[string]string)
var lists = make(map[string]*[]string)

func UpdateList(listkey string, val []string) int {
	var count int

	l, ok := lists[listkey]
	if !ok {
		nl := &[]string{} // new list
		*nl = append(*nl, val...)
		lists[listkey] = nl
		count = 1
	} else {
		*l = append(*l, val...)
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

		partcount, parts := RESPDecoder(n, in)
		// for _, v := range parts {
		// 	fmt.Printf("%q\n", v)
		// }

		switch strings.ToLower(parts[1]) {

		case "ping":
			out = RESPEncoder("PONG", 0)
			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "echo":
			out = RESPEncoder(parts[4], 1)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "get":
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

		case "set":
			key := parts[4]
			val := parts[6]

			vars[key] = val

			// parts has an element at index 8 and index 10
			// if parts has expiry args sent -> only then setup expiry
			if partcount == 5 {
				err := SetupExpiry(parts[8], parts[10], key)
				fmt.Printf("Error setting up expiry for key (%s): %s\n", key, err.Error())
			}

			out = RESPEncoder("OK", 0)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "rpush":
			listkey := parts[4]
			var vals []string

			// val count = part count - 2
			valcount := partcount - 2
			i := 6
			for valcount > 0 {
				vals = append(vals, parts[i])
				i += 2
				valcount--
			}

			listsize := UpdateList(listkey, vals)

			out = RESPEncoder(strconv.Itoa(listsize), 2)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}
		}
	}
}

func RESPDecoder(n int, in []byte) (int, []string) {
	com := string(in[:n])

	// print incoming bytes -> testing
	// fmt.Printf("%q\n", in[:n])

	parts := strings.Split(com, "\r\n")
	count, err := strconv.ParseInt(parts[0][1:], 10, 0)
	if err != nil {
		fmt.Println("Error parsing RESP arg count into an integer: %s\n", err.Error())
		return 0, []string{}
	}

	// prevent the last empty string from being passed as a part of the RESP command
	return int(count), parts[1 : len(parts)-1]
}

// simple string -> +{string}\r\n -> 0
// bulk string -> {string_len}\r\n{string}\r\n -> 1
// RESP integer -> :{integer (sent as a string)}\r\n -> 2
func RESPEncoder(res string, t int) []byte {
	var s string

	switch t {
	case 0:
		s = fmt.Sprintf("+%s\r\n", res)
	case 2:
		s = fmt.Sprintf(":%s\r\n", res)
	default: // t == 1, bulk string as default
		if len(res) == 0 {
			s = "$-1\r\n"
		} else {
			s = fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(res)), res)
		}
	}

	// if t == 0 {
	// 	s = fmt.Sprintf("+%s\r\n", res)
	// } else if t == 2 {
	// 	s = fmt.Sprintf(":%s\r\n", res)
	// } else if t == 1 {
	// 	if len(res) == 0 {
	// 		s = "$-1\r\n"
	// 	} else {
	// 		s = fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(res)), res)
	// 	}
	// }

	return []byte(s)
}

func SetupExpiry(t string, ds string, key string) error {
	// ex -> second
	// px -> millisecond
	var m time.Duration

	if strings.ToLower(t) == "px" {
		m = time.Millisecond
	} else if strings.ToLower(t) == "ex" {
		m = time.Second
	} else {
		return errors.New("Unknown expiry type")
	}

	d, err := strconv.ParseInt(ds, 10, 64)
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing the duration into an integer: %s\n", err.Error()))
	}

	go Expire(time.Duration(d)*m, key)
	return nil
}

func Expire(t time.Duration, key string) {
	time.Sleep(t)
	delete(vars, key)
}
