package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const (
	orderPrefix = "/order"
)

func ProcessOrder(wg *sync.WaitGroup) {
	var source, sink *Pos
	for {
		source = generateMapPoint()
		sink = generateMapPoint()
		if !cmp(*source, *sink) {
			break
		}
	}
	fmt.Println(source, sink)
	orderID := sendOrder(*source, *sink)
	distance := dis(*sink, *source)
	distanceDuration := time.Duration(distance) * time.Second
	fmt.Printf("order Id : %d\n", orderID)
	wg.Done()
	time.Sleep(orderBaseDuration + distanceDuration)
	fmt.Println(orderBaseDuration + distanceDuration)
	overOrder(orderID)
}

func ProcessOrderWithoutWG() {
	var source, sink *Pos
	for {
		source = generateMapPoint()
		sink = generateMapPoint()
		if !cmp(*source, *sink) {
			break
		}
	}
	fmt.Println(source, sink)
	orderID := sendOrder(*source, *sink)
	distance := dis(*sink, *source)
	distanceDuration := time.Duration(distance) * time.Second
	fmt.Printf("order Id : %d\n", orderID)
	time.Sleep(orderBaseDuration + distanceDuration)
	fmt.Println(orderBaseDuration + distanceDuration)
	overOrder(orderID)
}

func sendOrder(source, sink Pos) int {
	content, err := doPut(endpoint, fmt.Sprintf("%s?fromx=%f&fromy=%f&tox=%f&toy=%f", orderPrefix, source.x, source.y, sink.x, sink.y))
	if err != nil {
		fmt.Println(err)
		return -1
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(content, &data)
	fmt.Println(data)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	id, ok := data["Id"].(float64)
	if !ok {
		return -1
	}
	return int(id)
}

func overOrder(id int) error {
	err := doDelete(endpoint, fmt.Sprintf("%s/%d", orderPrefix, id))
	if err != nil {
		fmt.Println(err)
	}
	return err
}
