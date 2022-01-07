package main

import (
	"fmt"
	"sync"
)

type Demo struct {
	Param
	wg sync.WaitGroup
}

func (d *Demo) initDemo(begin, end int) {
	if end > 0 {
		fmt.Println(begin, end)
		for i := begin; i < end; i++ {
			overOrder(i)
		}
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
