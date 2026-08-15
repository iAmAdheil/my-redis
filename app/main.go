package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
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
		in := make([]byte, 1000)
		var out []byte

		n, err := conn.Read(in)
		if err != nil {
			fmt.Printf("Error reading from connection: %s\n", err.Error())
			break
		}

		base, args, err := RESPDecoder(in[:n])
		if err != nil {
			// do something
		}
		com, err := GetCom(base, args)
		if err != nil {
			// do something
		}
		// log sent over request
		// for _, v := range parts {
		// 	fmt.Printf("%q\n", v)
		// }

		switch base {

		case "ping":
			out = RESPEncoder([]string{"PONG"}, 0)
			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "echo":
			res := []string{parts[3]}
			out = RESPEncoder(res, 1)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "get":
			key := parts[3]
			res := []string{}
			val, ok := vars[key]
			if !ok {
				res = append(res, "")
			} else {
				res = append(res, val)
			}

			out = RESPEncoder(res, 1)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "set":
			key := parts[3]
			val := parts[5]

			vars[key] = val

			// parts has an element at index 8 and index 10
			// if parts has expiry args sent -> only then setup expiry
			if partcount == 5 {
				err := SetupExpiry(parts[7], parts[9], key)
				if err != nil {
					fmt.Printf("Error setting up expiry for key (%s): %s\n", key, err.Error())
				}
			}

			res := []string{"OK"}
			out = RESPEncoder(res, 0)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "rpush":
			listkey := parts[3]
			var vals []string

			// val count = part count - 2
			valcount := partcount - 2
			i := 5
			for valcount > 0 {
				vals = append(vals, parts[i])
				i += 2
				valcount--
			}

			listsize := AddToList(listkey, vals, 0)

			res := []string{strconv.Itoa(listsize)}
			out = RESPEncoder(res, 2)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "lrange":
			key := parts[3]
			l, err := strconv.ParseInt(parts[5], 10, 0)
			if err != nil {
				fmt.Printf("Error parsing the start index into an integer: %s\n", err.Error())
			}
			r, err := strconv.ParseInt(parts[7], 10, 0)
			if err != nil {
				fmt.Printf("Error parsing the stop index into an integer: %s\n", err.Error())
			}

			res := GetListRange(key, int(l), int(r))
			out = RESPEncoder(res, 3)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "lpush":
			key := parts[3]

			vals := []string{}
			valCount := partcount - 2
			i := 5

			for valCount > 0 {
				vals = append([]string{parts[i]}, vals...)
				i += 2
				valCount--
			}

			listsize := AddToList(key, vals, 1)

			res := []string{strconv.Itoa(listsize)}
			out = RESPEncoder(res, 2)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "llen":
			key := parts[3]
			listsize := GetListLen(key)

			res := []string{strconv.Itoa(listsize)}
			out = RESPEncoder(res, 2)

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}

		case "lpop":
			key := parts[3]
			count := 1

			if len(parts) == 6 {
				pc, err := strconv.ParseInt(parts[5], 10, 0) //parsed count
				if err == nil {
					count = int(pc)
				}
			}

			res := []string{}
			s, err := DeleteFromList(key, count, 1)
			if err == nil {
				res = append(res, s...)
			}

			var out []byte
			if len(res) <= 1 {
				// bulk string
				out = RESPEncoder(res, 1)
			} else {
				// bulk string list
				out = RESPEncoder(res, 3)
			}

			if _, err := conn.Write(out); err != nil {
				fmt.Printf("Error writing into connection: %s\n", err.Error())
			}
		}
	}
}
