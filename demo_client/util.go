package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"time"
)

const (
	eps = 1e-4

	locationProfix = "/location"
)

type Pos struct {
	x float64
	y float64
}

func cmp(a, b Pos) bool {
	return math.Abs(a.x-b.x) < eps && math.Abs(a.y-b.y) < eps
}

func sqr(x float64) float64 {
	return x * x
}

func dis(a, b Pos) float64 {
	return math.Sqrt(sqr(a.x-b.x) + sqr(a.y-b.y))
}

func generateRandomNumber(start int, end int, count int) []int {
	if end-start+1 < count {
		return nil
	}

	ret := make([]int, 0, count)
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < count; i++ {
		var toAdd int
		for {
			toAdd = generator.Intn(end-start+1) + start
			ok := true
			for _, item := range ret {
				if item == toAdd {
					ok = false
					break
				}
			}
			if ok {
				break
			}
		}
		ret = append(ret, toAdd)
	}
	return ret
}

func generateMapPoint() *Pos {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	content, err := doGet(endpoint, locationProfix)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(content, &data)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	x, ok := data["X"].(float64)
	if !ok {
		fmt.Println(err)
		return nil
	}
	y, ok := data["Y"].(float64)
	if !ok {
		fmt.Println(err)
		return nil
	}
	fmt.Println("One Pos", x, y)
	return &Pos{x: x, y: y}
}

func doRequest(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func doGet(endpoint, prefix string) ([]byte, error) {
	url := endpoint + prefix
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return doRequest(req)
}

func doPut(endpoint, prefix string) ([]byte, error) {
	url := endpoint + prefix
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return nil, err
	}
	return doRequest(req)
}

func doDelete(endpoint, prefix string) ([]byte, error) {
	url := endpoint + prefix
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return content, nil
}
