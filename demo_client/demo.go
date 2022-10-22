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

		var oneTime sync.WaitGroup
		oneTime.Add(1)
		go func() {
			time.Sleep(d.intervalTime)
			oneTime.Done()
		}()
		d.OrderDemo()
		d.changeShifts()
		oneTime.Wait()
	}
	time.Sleep(30 * time.Second)
}

func (d *Demo) initDemo(begin, end int) {
	if end > 0 {
		closeOrder(begin, end)
	}
	if getOffWork > 0 {
		d.getOffWorkInit()
	}
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
