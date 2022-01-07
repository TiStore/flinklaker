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

	orderBaseDuration = 5 * time.Second
	demoDuration      = 20
	intervalTime      = 2 * time.Second

	endpoint = "http://localhost:7998"
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
	getOffWorkInit()
	goOnWork(initOnWorkNum)
}

func changeShifts() {
	getOffWork(changeShiftsNum)
	goOnWork(changeShiftsNum)
}

func OrderDemo() {
	wg.Add(orderNum)
	for i := 0; i < orderNum; i++ {
		go ProcessOrder(&wg)
	}
	wg.Wait()
}
