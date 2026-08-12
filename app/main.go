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

// # Imp points (@iAmAdheil) -> pls take a look later
// - arg validation should happen during decoding -> string to int conversions should not happen within my com handlers

var vars = make(map[string]string)
var lists = make(map[string]*[]string)

// dir -> 0 for append
// dir -> 1 for prepend
func UpdateList(listkey string, val []string, dir int) int {
	var count int

	l, ok := lists[listkey]
	if !ok {
		l = &[]string{} // new list
		lists[listkey] = l
	}

	switch dir {
	case 0:
		*l = append(*l, val...)

	case 1:
		*l = append(val, *l...)
	}

	count = len(*l)
	return count
}

func GetListRange(key string, l, r int) []string {
	res := []string{}

	list, ok := lists[key]
	if !ok {
		return res
	}
	listsize := len(*list)

	if l < 0 {
		l = max(0, listsize-(-1*l))
	}
	if r < 0 {
		r = max(0, listsize-(-1*r))
	}

	// max index for the list
	rmax := listsize - 1

	if l > r || l > rmax {
		return res
	}

	for i := l; i <= min(r, rmax); i++ {
		res = append(res, (*list)[i])
	}

	return res
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
		in := make([]byte, 1000)
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

			listsize := UpdateList(listkey, vals, 0)

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
			list, ok := lists[key]
			if !ok {
				list = &[]string{}
				lists[key] = list
			}

			vals := []string{}
			valCount := partcount - 2
			i := 5

			for valCount > 0 {
				vals = append([]string{parts[i]}, vals...)
				i += 2
				valCount--
			}

			listsize := UpdateList(key, vals, 1)

			res := []string{strconv.Itoa(listsize)}
			out = RESPEncoder(res, 2)

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
// bulk string -> ${string_len}\r\n{string}\r\n -> 1
// RESP integer -> :{integer (sent as a string)}\r\n -> 2
// bulk string list -> *{res count} ... ${string_len}\r\n{string}\r\n -> 1=3
func RESPEncoder(res []string, t int) []byte {
	var s string

	switch t {
	case 0:
		s = fmt.Sprintf("+%s\r\n", res[0])
	case 2:
		s = fmt.Sprintf(":%s\r\n", res[0])
	case 3:
		s = fmt.Sprintf("*%s\r\n", strconv.Itoa(len(res)))
		for _, v := range res {
			s += fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(v)), v)
		}
	default: // t == 1, bulk string as default
		if len(res) == 0 || len(res[0]) == 0 {
			s = "$-1\r\n"
		} else {
			v := res[0]
			s = fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(v)), v)
		}
	}

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
		return fmt.Errorf("Error parsing the duration into an integer: %s\n", err.Error())
	}

	go Expire(time.Duration(d)*m, key)
	return nil
}

func Expire(t time.Duration, key string) {
	time.Sleep(t)
	delete(vars, key)
}
