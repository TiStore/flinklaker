package main

import (
	"fmt"
)

const (
	carPrefix = "/car"
)

func (d *Demo) getOffWorkInit() {
	for i := 0; i < d.pointNum; i++ {
		letCarGetOffWorkByID(i)
	}
}

func (d *Demo) getOffWork(n int) {
	ids := generateRandomNumber(1, d.pointNum, n)
	for _, id := range ids {
		letCarGetOffWorkByID(id)
	}
}

func (d *Demo) goOnWork(n int) {
	ids := generateRandomNumber(1, d.pointNum, n)
	for _, id := range ids {
		_, err := letCarGoToWorkByID(id)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func letCarGetOffWorkByID(id int) bool {
	_, err := doDelete(endpoint, fmt.Sprintf("%s/%d", carPrefix, id))
	if err != nil {
		fmt.Println(err)
	}
	return true
}

func letCarGoToWorkByID(id int) ([]byte, error) {
	pos := generateMapPoint()
	return doPut(endpoint, fmt.Sprintf("%s/%d?x=%f&y=%f", carPrefix, id, pos.x, pos.y))
}
