package main

import (
	"fmt"
	"sync"
	"time"
)

type Demo struct {
	Param
	wg sync.WaitGroup
}

func (d *Demo) doDemo() {
	d.initDemo(orderBegin, orderEnd)

	for page := 0; page < d.demoTimes; page++ {
		fmt.Println(page)

		d.OrderDemo()

		d.changeShifts()
		time.Sleep(d.intervalTime)
	}
	time.Sleep(15 * time.Second)
}

func (d *Demo) initDemo(begin, end int) {
	if end > 0 {
		closeOrder(begin, end)
	}
	d.getOffWorkInit()
	d.goOnWork(d.initOnWorkNum)
}

func (d *Demo) changeShifts() {
	d.getOffWork(d.changeShiftsNum)
	d.goOnWork(d.changeShiftsNum)
}

func (d *Demo) OrderDemo() {
	d.wg.Add(d.orderNum)
	for i := 0; i < d.orderNum; i++ {
		go d.ProcessOrder(&d.wg)
	}
	d.wg.Wait()
}
