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
	disDura           int

	demoTimes    int
	intervalTime time.Duration

	distanceLimit float64

	delayRequest bool
}

var orderBegin, orderEnd int
var action int

func init() {
	flag.IntVar(&action, "t", 0, "")

	flag.IntVar(&orderBegin, "ob", 0, "")
	flag.IntVar(&orderEnd, "oe", 0, "")
}

func main() {

	flag.Parse()

	fmt.Println(action, orderBegin, orderEnd)
	if action == 1 {
		closeOrder(orderBegin, orderEnd)
		return
	} else if action == 0 {
		FirstDemo()
	} else if action == 2 {
		OneOrderDemo()
	}
}

func FirstDemo() {
	demo := &Demo{
		Param: Param{
			pointNum:          50,
			initOnWorkNum:     12,
			changeShiftsNum:   4,
			orderNum:          3,
			orderBaseDuration: 3 * time.Second,
			disDura:           10,
			demoTimes:         10,
			intervalTime:      5 * time.Second,
			distanceLimit:     0.006,
		},
	}
	demo.doDemo()
}

func OneOrderDemo() {
	demo := &Demo{
		Param: Param{
			pointNum:          50,
			initOnWorkNum:     15,
			changeShiftsNum:   0,
			orderNum:          1,
			orderBaseDuration: 30 * time.Second,
			disDura:           39,
			demoTimes:         1,
			intervalTime:      30 * time.Second,
			distanceLimit:     0.006,
		},
	}
	demo.doDemo()
}
