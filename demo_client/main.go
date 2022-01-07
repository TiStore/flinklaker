package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

const (
	PointNum        = 36668
	InitOnWorkNum   = 2000
	ChangeShiftsNum = 300
	OrderNum        = 50

	OrderBaseDuration = 5 * time.Second
	DemoTimes         = 20
	IntervalTime      = 2 * time.Second

	endpoint = "http://localhost:7998"
)

type Param struct {
	pointNum          int
	initOnWorkNum     int
	changeShiftsNum   int
	orderNum          int
	orderBaseDuration time.Duration
	demoTimes         int
	intervalTime      time.Duration

	distanceLimit float64
}

var orderBegin, orderEnd int

func main() {

	rand.Seed(time.Now().UnixNano())

	flag.IntVar(&orderBegin, "ob", 0, "")
	flag.IntVar(&orderEnd, "oe", 0, "")
	FirstDemo()
}

func FirstDemo() {
	demo := &Demo{
		Param: Param{
			pointNum:          50,
			initOnWorkNum:     10,
			changeShiftsNum:   3,
			orderNum:          3,
			orderBaseDuration: 5 * time.Second,
			demoTimes:         10,
			intervalTime:      2 * time.Second,
			distanceLimit:     0.0001,
		},
	}
	demo.initDemo(0, 0)

	for page := 0; page < demo.demoTimes; page++ {
		fmt.Println(page)

		demo.OrderDemo()

		demo.changeShifts()
		time.Sleep(demo.intervalTime)
	}
}
