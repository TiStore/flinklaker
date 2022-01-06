package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

const (
	pointNum        = 36668
	initOnWorkNum   = 2000
	changeShiftsNum = 300
	orderNum        = 50

	orderBaseDuration = 1 * time.Second
	demoDuration      = 20
	intervalTime      = 1 * time.Second

	endpoint = "http://localhost:8000"
)

func main() {

	initDemo()

	for page := 0; page < demoDuration; page++ {
		fmt.Println(page)

		OrderDemo()

		changeShifts()
		time.Sleep(intervalTime)
	}
}

func initDemo() {
	getOffWork(pointNum)
	goOnWork(initOnWorkNum)
}

func changeShifts() {
	getOffWork(changeShiftsNum)
	goOnWork(changeShiftsNum)
}

func OrderDemo() {
	wg.Add(orderNum)
	for i := 0; i < orderNum; i++ {
		go ProcessOrder(wg)
	}
	wg.Wait()
}
