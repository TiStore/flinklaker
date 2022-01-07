package main

import (
	"flag"
	"fmt"
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
var onlyCloseOrder bool

func main() {

	flag.BoolVar(&onlyCloseOrder, "close", false, "")

	flag.IntVar(&orderBegin, "ob", 0, "")
	flag.IntVar(&orderEnd, "oe", 0, "")

	if onlyCloseOrder {
		closeOrder(orderBegin, orderEnd)
		return
	}

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
	demo.initDemo(orderBegin, orderEnd)

	for page := 0; page < demo.demoTimes; page++ {
		fmt.Println(page)

		demo.OrderDemo()

		demo.changeShifts()
		time.Sleep(demo.intervalTime)
	}
}
