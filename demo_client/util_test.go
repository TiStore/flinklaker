package main

import (
	"fmt"
	"testing"
)

func TestLocation(t *testing.T) {
	pos := generateMapPoint()
	fmt.Printf("test %v\n", *pos)
}
