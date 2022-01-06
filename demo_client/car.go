package main

import "fmt"

const (
	carPrefix = "/car"
)

func getOffWork(n int) {
	ids := generateRandomNumber(1, pointNum, n)
	for _, id := range ids {
		err := letCarGetOffWorkByID(id)
		fmt.Println(err)
	}
}

func goOnWork(n int) {
	ids := generateRandomNumber(1, pointNum, n)
	for _, id := range ids {
		_, err := letCarGoToWorkByID(id)
		fmt.Println(err)
	}
}

func letCarGetOffWorkByID(id int) error {
	return doDelete(endpoint, fmt.Sprintf("%s/%d", carPrefix, id))
}

func letCarGoToWorkByID(id int) ([]byte, error) {
	return doPut(endpoint, fmt.Sprintf("%s/%d", carPrefix, id))
}
