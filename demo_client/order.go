package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	orderPrefix = "/order"
)

func (d *Demo) ProcessOrder(wg *sync.WaitGroup) {
	var source, sink *Pos
	for i := 0; i < 20; i++ {
		source = generateMapPoint()
		sink = generateMapPoint()
		if source == nil || sink == nil {
			continue
		}
		if dis(*source, *sink) > d.distanceLimit {
			break
		}
	}
	fmt.Println(source, sink)
	orderID := sendOrder(*source, *sink)
	distance := dis(*sink, *source)
	distanceDuration := time.Duration(distance*5) * time.Second
	fmt.Printf("order Id : %d\n", orderID)
	wg.Done()
	for i := 0; i < 10; i++ {
		time.Sleep(d.orderBaseDuration + distanceDuration)
		if overOrder(orderID) {
			break
		}
	}
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

func overOrder(id int) bool {
	content, err := doDelete(endpoint, fmt.Sprintf("%s/%d", orderPrefix, id))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(content))
	return strings.HasPrefix(string(content), fmt.Sprintf("Finish order (id:%v) Success", id))
}
