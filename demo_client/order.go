package main

import (
	"fmt"
	"time"
)

const (
	orderPrefix = "/order"
)

func ProcessDemo() {
	var source, sink Pos
	for {
		source = generateMapPoint()
		sink = generateMapPoint()
		if !cmp(source, sink) {
			break
		}
	}
	orderID := sendOrder(source, sink)
	distance := dis(sink, source)
	distanceDuration := time.Duration(distance) * time.Second
	time.Sleep(orderBaseDuration + distanceDuration)
	overOrder(orderID)
}

func sendOrder(source, sink Pos) int {
	_, err := doPut(endpoint, fmt.Sprintf("%s?fromx=%f&fromy=%f&tox=%f&toy=%f", orderPrefix, source.x, source.y, sink.x, sink.y))
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return 0
}

func overOrder(id int) error {
	err := doDelete(endpoint, fmt.Sprintf("%s/%d", orderPrefix, id))
	if err != nil {
		fmt.Println(err)
	}
	return err
}
