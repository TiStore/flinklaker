package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"xorm.io/core"
)

type Car struct {
	Id         int64  `xorm:"pk autoincr"` //指定主键并自增
	LocationX  int64  `xorm:"'location_x'"`
	LocationY  int64  `xorm:"'location_y'"`
	Status     string `xorm:"'status'"`
	CreateTime string `xorm:"created 'create_time'"`
	UpdateTime string `xorm:"updated 'update_time'"`
}

func (c *Car) TableName() string {
	return "cars"
}

func NewCar(id, x, y int64) *Car {
	return &Car{
		Id:         id,
		LocationX:  x,
		LocationY:  y,
		Status:     "idle",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		UpdateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
}

func beginTxnAndGetCarByID(carID int64) (*core.Tx, *Car, error) {
	txn, err := engine.DB().Begin()
	if err != nil {
		return nil, nil, err
	}
	raws := txn.QueryRow("select status,location_x,location_y from cars where id=?", carID)
	var car Car
	err = raws.Scan(&car.Status, &car.LocationX, &car.LocationY)
	if err != nil {
		return txn, nil, err
	}
	return txn, &car, nil
}

// PUT /car/{id}?x=${x}&y=${y}
// create a car or online a car
// Example: curl -X PUT "http://localhost:8000/car/1?x=12&y=13"
func PutNewCar(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	carID, err := strconv.ParseInt(params["id"], 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}

	values := r.URL.Query()

	loX, err := strconv.ParseInt(values.Get("x"), 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}
	loY, err := strconv.ParseInt(values.Get("y"), 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}
	car := NewCar(carID, loX, loY)
	tx, oldCar, err := beginTxnAndGetCarByID(carID)
	defer tx.Commit()
	if err == nil {
		if oldCar.Status == "offline" {
			res, err := tx.Exec("update cars set status='idle',update_time=? where id=?", car.UpdateTime, carID)
			w.Write([]byte(fmt.Sprintf("Active car(ID=%v) set status='idle' success! %v,%v", carID, res, err)))
			return
		} else {
			w.Write([]byte(fmt.Sprintf("Active car(id:%v) Failed,invalid status: %v\n", carID, oldCar.Status)))
			return
		}
	}

	n, err := engine.Insert(car)
	_, err = w.Write([]byte(fmt.Sprintf("online a car success! ID:%v,location(%v,%v),dd %v,err %v\n", carID, loX, loY, n, err)))
}

// DELETE /car/{id}
// offline a car
// Example: curl -X DELETE "http://localhost:8000/car/1"
func DeleteCar(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	carID, err := strconv.ParseInt(params["id"], 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}
	tx, oldCar, err := beginTxnAndGetCarByID(carID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Commit()

	if oldCar.Status != "idle" {
		w.Write([]byte(fmt.Sprintf("Offline car(ID=%v) failed since the status is %v!", carID, oldCar.Status)))
		return
	}

	_, err = tx.Exec("update cars set status='offline' where id=?", carID)
	_, _ = w.Write([]byte(fmt.Sprintf("offline a car success! ID:%v,err %v\n", carID, err)))
}
