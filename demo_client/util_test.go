package main

import (
	"fmt"
	"testing"
)

func TestGenerateRandomNumber(t *testing.T) {
	ret := generateRandomNumber(0, 100, 20)
	fmt.Println(ret)
}

func TestLocation(t *testing.T) {
	pos := generateMapPoint()
	fmt.Printf("test %v\n", *pos)
}
