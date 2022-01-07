package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"xorm.io/core"
)

type Order struct {
	Id         int64   `xorm:"pk autoincr 'order_id'"` //指定主键并自增
	CarID      int64   `xorm:"'car_id'"`
	FromX      float64 `xorm:"'from_x'"`
	FromY      float64 `xorm:"'from_y'"`
	ToX        float64 `xorm:"'to_x'"`
	ToY        float64 `xorm:"'to_y'"`
	Status     string  `xorm:"'status'"`
	CreateTime string  `xorm:"created 'create_time'"`
	UpdateTime string  `xorm:"updated 'update_time'"`
}

func (c *Order) TableName() string {
	return "orders"
}

func (o *Order) initTime() {
	o.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	o.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
}

func beginTxnAndOrderByID(orderID int64) (*core.Tx, *Order, error) {
	txn, err := engine.DB().Begin()
	if err != nil {
		return nil, nil, err
	}
	raws := txn.QueryRow("select status,from_x,from_y,to_x,to_y,car_id from orders where order_id=?", orderID)
	order := Order{Id: orderID}
	err = raws.Scan(&order.Status, &order.FromX, &order.FromY, &order.ToX, &order.ToY, &order.CarID)
	if err != nil {
		return txn, nil, err
	}
	return txn, &order, nil
}

// PUT /order?fromx=?&fromy=?&tox=?toy=?
// a new order
// Example: curl -X PUT "http://localhost:8000/order?fromx=1&fromy=2&tox=12.2&toy=13.3"
// {"Id":3,"CarID":0,"FromX":1,"FromY":2,"ToX":12.2,"ToY":13.3,"Status":"waiting","CreateTime":"2022-01-06 21:40:23","UpdateTime":"2022-01-06 21:40:23"}
func PutNewOrder(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	order := &Order{}
	var err error
	defer func() {
		if err != nil {
			writeError(w, err)
			return
		}
	}()
	order.FromX, err = strconv.ParseFloat(values.Get("fromx"), 64)
	if err != nil {
		err = fmt.Errorf("parse arg(fromx) faild:%v", err)
		return
	}
	order.FromY, err = strconv.ParseFloat(values.Get("fromy"), 64)
	if err != nil {
		err = fmt.Errorf("parse arg(fromy) faild:%v", err)
		return
	}

	order.ToX, err = strconv.ParseFloat(values.Get("tox"), 64)
	if err != nil {
		err = fmt.Errorf("parse arg(tox) faild:%v", err)
		return
	}
	order.ToY, err = strconv.ParseFloat(values.Get("toy"), 64)
	if err != nil {
		err = fmt.Errorf("parse arg(toy) faild:%v", err)
		return
	}
	order.initTime()
	order.Status = "waiting"
	var res sql.Result
	res, err = engine.Exec("insert into orders set from_x=?,from_y=?,to_x=?,to_y=?,status=?,create_time=?,update_time=?",
		order.FromX, order.FromY, order.ToX, order.ToY, order.Status, order.CreateTime, order.UpdateTime)
	if err != nil {
		err = fmt.Errorf("insert  db(order) failed:%v", err)
		return
	}
	order.Id, err = res.LastInsertId()
	if err != nil || order.Id == 0 {
		err = fmt.Errorf("insert  db(order) failed:%v,lastinsertID:%d", err, order.Id)
		return
	}
	writeJson(w, order)
}

// DELETE /order/{orderID}
// Finish an order
// Example: curl -X DELETE "http://localhost:8000/order/1"
func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	orderID, err := strconv.ParseInt(params["id"], 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}
	tx, order, err := beginTxnAndOrderByID(orderID)

	defer func() {
		if err != nil {
			if tx != nil {
				tx.Rollback()
			}
			writeError(w, err)
		}
	}()
	if err != nil {
		err = fmt.Errorf("Get Order(id:%v) Failed:%v", orderID, err)
		return
	}
	if order.Status != "running" {
		err = fmt.Errorf("Finish order(id:%v) Failed,invalid status: %v\n", orderID, order.Status)
		return
	}
	var res sql.Result
	// Set car to idle & update the location
	res, err = tx.Exec("update cars set status='idle',location_x=?,location_y=? where id=?", order.ToX, order.ToY, order.CarID)
	if err != nil {
		err = fmt.Errorf("update db(cars) failed:%v", err)
		return
	}
	var rows int64
	rows, err = res.RowsAffected()
	if err != nil || rows == 0 {
		err = fmt.Errorf("update db(cars) failed:%v,affected rows:%v", err, rows)
		return
	}
	_, err = tx.Exec("update orders set status='finished' where id=?", orderID)
	if err != nil {
		return
	}

	err = tx.Commit()
	w.Write([]byte(fmt.Sprintf("Finish order (id:%v) Success,car(id:%v) is idle now", orderID, order.CarID)))

}

type NearCars struct {
	cars    []int64
	orderID int64
}

func checkAndRunOrders() error {
	res, err := engine.Query("select order_id,cars from nearcars where order_id in(select order_id from orders where status='waiting') order by order_id;")
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, data := range res {
		var order NearCars
		order.orderID, err = strconv.ParseInt(string(data["order_id"]), 0, 64)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data["cars"], &order.cars)
		if err != nil {
			return err
		}

		log.Printf("nearcarsinfo:%+v\n", order)
		wg.Add(1)
		go func() {
			//give the order an car.
			err = runningTheOrder(&order)
			log.Printf("Finished to running the order:%+v,err:%v\n", order, err)
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}

func runningTheOrder(orderCars *NearCars) error {
	tx, order, err := beginTxnAndOrderByID(orderCars.orderID)
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	if order.Status != "waiting" {
		log.Printf("This order is not waiting:%+v\n", order)
		return nil
	}
	for _, carID := range orderCars.cars {
		var res sql.Result
		res, err = tx.Exec("update cars set status='running' where id=? and status='idle'", carID)
		if err != nil {
			return err
		}

		affectedRow, terr := res.RowsAffected()
		if terr != nil || affectedRow != 1 {
			fmt.Printf("unmatched order:%+v,carIID:%d,affactedrow:%d\n", order, carID, affectedRow)
			continue
		}
		order.CarID = carID
		order.Status = "running"
		break
	}
	order.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	var reso sql.Result
	reso, err = tx.Exec("update orders set status=?,car_id=?,update_time=? where order_id=?", order.Status, order.CarID, order.UpdateTime, order.Id)
	if err != nil {
		return err
	}
	affacted, _ := reso.RowsAffected()
	if affacted != 1 {
		fmt.Printf("affacted row for update orders should not be :%d\n", affacted)
	}
	err = tx.Commit()
	return err
}
