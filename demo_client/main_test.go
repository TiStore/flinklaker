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

		getOffWorkTest(4)
		goOnWorkTest(4)
		time.Sleep(intervalTime)
	}
}

func getOffWorkTest(n int) {
	ids := generateRandomNumber(1, 20, n)
	for _, id := range ids {
		err := letCarGetOffWorkByID(id)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func goOnWorkTest(n int) {
	ids := generateRandomNumber(1, 20, n)
	for _, id := range ids {
		_, err := letCarGoToWorkByID(id)
		fmt.Println(err)
	}
}
