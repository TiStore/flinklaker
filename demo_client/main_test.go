package main

import (
	"fmt"
	"testing"
	"time"
)

func TestMiniDemo(t *testing.T) {
	ids := generateRandomNumber(1, 20, 10)
	for _, id := range ids {
		err := letCarGetOffWorkByID(id)
		if err != nil {
			fmt.Println(err)
		}
	}
	goOnWork(20)

	for page := 0; page < 2; page++ {
		fmt.Println(page)

		for i := 0; i < 2; i++ {
			go ProcessOrder()
		}

		getOffWork(4)
		goOnWork(4)
		time.Sleep(intervalTime)
	}
}
