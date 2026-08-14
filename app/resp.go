package main

import (
	"fmt"
	"strconv"
	"strings"
)

func RESPDecoder(n int, in []byte) (int, []string) {
	com := string(in[:n])

	// print incoming bytes -> testing
	// fmt.Printf("%q\n", in[:n])

	parts := strings.Split(com, "\r\n")
	count, err := strconv.ParseInt(parts[0][1:], 10, 0)
	if err != nil {
		fmt.Printf("Error parsing RESP arg count into an integer: %s\n", err.Error())
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
