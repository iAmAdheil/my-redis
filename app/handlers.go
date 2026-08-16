package main

import (
	"fmt"
	"strconv"
)

func (com *Com) ping() []byte {
	return RESPEncoder([]string{"PONG"}, Simple)
}

func (com *Com) echo() []byte {
	s := com.Args["echothis"][0]

	return RESPEncoder([]string{s}, Bulk)
}

func (com *Com) get() []byte {
	key := com.Args["key"][0]

	res := []string{}

	vmu.Lock()
	val, ok := vars[key]
	defer vmu.Unlock()

	if !ok {
		res = append(res, "")
	} else {
		res = append(res, val)
	}

	return RESPEncoder(res, Bulk)
}

func (com *Com) set() []byte {
	key := com.Args["key"][0]
	val := com.Args["value"][0]

	vmu.Lock()
	vars[key] = val
	vmu.Unlock()

	// parts has an element at index 8 and index 10
	// if parts has expiry args sent -> only then setup expiry
	if len(com.Extras) > 0 {
		expiryType := com.Extras["expiryType"][0]
		dur := com.Extras["duration"][0]

		err := SetupExpiry(expiryType, dur, key)
		if err != nil {
			fmt.Printf("Error setting up expiry for key (%s): %s\n", key, err.Error())
		}
	}

	return RESPEncoder([]string{"OK"}, Simple)
}

func (com *Com) rpush() []byte {
	listkey := com.Args["listkey"][0]
	values := com.Args["values"] // array of values

	listsize := AddToList(listkey, values, 0)

	return RESPEncoder([]string{strconv.Itoa(listsize)}, Int)
}

func (com *Com) lrange() []byte {
	listkey := com.Args["listkey"][0]
	ls := com.Args["left"][0]
	rs := com.Args["right"][0]

	l, err := strconv.ParseInt(ls, 10, 0)
	if err != nil {
		fmt.Printf("Error parsing the start index into an integer: %s\n", err.Error())
	}
	r, err := strconv.ParseInt(rs, 10, 0)
	if err != nil {
		fmt.Printf("Error parsing the stop index into an integer: %s\n", err.Error())
	}

	out := GetListRange(listkey, int(l), int(r))
	return RESPEncoder(out, BulkList)
}

func (com *Com) lpush() []byte {
	listkey := com.Args["listkey"][0]
	values := com.Args["values"]

	listsize := AddToList(listkey, values, 1)

	out := []string{strconv.Itoa(listsize)}
	return RESPEncoder(out, Int)
}

func (com *Com) llen() []byte {
	listkey := com.Args["listkey"][0]
	listsize := GetListLen(listkey)

	out := []string{strconv.Itoa(listsize)}
	return RESPEncoder(out, Int)
}

func (com *Com) lpop() []byte {
	listkey := com.Args["listkey"][0]
	count, err := strconv.ParseInt(com.Args["count"][0], 10, 0)
	if err != nil {
		// do something
	}

	out := []string{}
	s, err := DeleteFromList(listkey, int(count), 1)
	if err == nil {
		out = append(out, s...)
	}

	if len(out) > 1 {
		// bulk list
		return RESPEncoder(out, BulkList)
	}
	// bulk string
	return RESPEncoder(out, Bulk)
}
