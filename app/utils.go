package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// # Imp points (@iAmAdheil) -> pls take a look later
// - arg validation should happen during decoding -> string to int conversions should not happen within my com handlers

var vars = make(map[string]string)
var vmu sync.Mutex = sync.Mutex{}

var lists = make(map[string]*[]string)
var lmu sync.RWMutex = sync.RWMutex{}

// dir -> 0 for append
// dir -> 1 for prepend
func AddToList(listkey string, val []string, dir int) int {
	var count int

	lmu.Lock()
	l, ok := lists[listkey]
	defer lmu.Unlock()

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

	lmu.Lock()
	l, ok := lists[listkey]
	defer lmu.Unlock()

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

func GetListLen(listkey string) int {
	lmu.RLock()
	l, ok := lists[listkey]
	defer lmu.RUnlock()

	if !ok {
		return 0
	}

	return len(*l)
}

func GetListRange(listkey string, l, r int) []string {
	res := []string{}

	lmu.RLock()
	list, ok := lists[listkey]
	defer lmu.RUnlock()

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

func SetupExpiry(et string, dur string, key string) error {
	// et -> expiry type
	// ex -> second
	// px -> millisecond
	var m time.Duration

	if strings.ToLower(et) == "px" {
		m = time.Millisecond
	} else if strings.ToLower(et) == "ex" {
		m = time.Second
	} else {
		return errors.New("Unknown expiry type")
	}

	d, err := strconv.ParseInt(dur, 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing the duration into an integer: %s\n", err.Error())
	}

	go Expire(time.Duration(d)*m, key)
	return nil
}

func Expire(t time.Duration, key string) {
	time.Sleep(t)

	vmu.Lock()
	delete(vars, key)
	vmu.Unlock()
}
