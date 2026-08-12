package main

import (
	"fmt"
	"testing"
)

func TestRand(t *testing.T) {
	s := []string{"1", "2", "3"}
	s = s[3:]

	fmt.Println(s)
}
