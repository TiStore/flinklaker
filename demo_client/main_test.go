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

	for page := 0; page < 1; page++ {

		wg.Add(2)
		for i := 0; i < 2; i++ {
			fmt.Println("i", i)
			go ProcessOrder(wg)
		}
		wg.Wait()

		getOffWorkTest(4)
		goOnWorkTest(4)
		time.Sleep(intervalTime)
	}
	time.Sleep(10 * time.Second)
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
