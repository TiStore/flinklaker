package main

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	orderPrefix = "/order"
)

func ProcessOrder() {
	var source, sink *Pos
	for {
		source = generateMapPoint()
		sink = generateMapPoint()
		if !cmp(*source, *sink) {
			break
		}
	}
	orderID := sendOrder(*source, *sink)
	distance := dis(*sink, *source)
	distanceDuration := time.Duration(distance) * time.Second
	fmt.Printf("order Id : %d\n", orderID)
	time.Sleep(orderBaseDuration + distanceDuration)
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
	if err != nil {
		fmt.Println(err)
		return -1
	}
	id, ok := data["Id"].(int)
	if !ok {
		return -1
	}
	return id
}

func overOrder(id int) error {
	err := doDelete(endpoint, fmt.Sprintf("%s/%d", orderPrefix, id))
	if err != nil {
		fmt.Println(err)
	}
	return err
}
