package main

import (
	"errors"
	"fmt"
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
func AddToList(listkey string, val []string, dir int) int {
	var count int

	l, ok := lists[listkey]
	if !ok {
		l = &[]string{} // new list
		lists[listkey] = l
	}

	switch dir {
	// rpush
	case 0:
		*l = append(*l, val...)
	// lpush
	case 1:
		*l = append(val, *l...)
	}

	count = len(*l)
	return count
}

func DeleteFromList(listkey string, count, dir int) ([]string, error) {
	s := []string{}

	l, ok := lists[listkey]
	if !ok {
		return []string{}, fmt.Errorf("Key not found")
	}

	listsize := len(*l)
	// set max removable elements if count > list size
	count = min(listsize, count)

	switch dir {
	// rpop
	case 0:
	// lpop
	case 1:
		for i := 0; i < count; i++ {
			s = append(s, (*l)[i])
		}
		*l = (*l)[count:]
	}

	return s, nil
}

func GetListLen(key string) int {
	l, ok := lists[key]
	if !ok {
		return 0
	}

	return len(*l)
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
