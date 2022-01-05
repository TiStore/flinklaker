package main

import (
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"xorm.io/core"
)

type Order struct {
	Id         int64  `xorm:"pk autoincr 'order_id'"` //指定主键并自增
	CarID      int64  `xorm:"'car_id'"`
	FromX      int64  `xorm:"'from_x'"`
	FromY      int64  `xorm:"'from_y'"`
	ToX        int64  `xorm:"'to_x'"`
	ToY        int64  `xorm:"'to_y'"`
	Status     string `xorm:"'status'"`
	CreateTime string `xorm:"created 'create_time'"`
	UpdateTime string `xorm:"updated 'update_time'"`
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
	err = raws.Scan(&order.Status, &order.FromX, &order.FromY, &order.ToX, &order.ToY, order.CarID)
	if err != nil {
		return txn, nil, err
	}
	return txn, &order, nil
}

// PUT /order?fromx=?&fromy=?&tox=?toy=?
// a new order
// Example: curl -X PUT "http://localhost:8000/order?fromx=1&fromy=2&tox=12&toy=13"
func PutNewOrder(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	order := &Order{}
	var err error
	order.FromX, err = strconv.ParseInt(values.Get("fromx"), 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}
	order.FromY, err = strconv.ParseInt(values.Get("fromy"), 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}

	order.ToX, err = strconv.ParseInt(values.Get("tox"), 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}
	order.ToY, err = strconv.ParseInt(values.Get("toy"), 0, 64)
	if err != nil {
		writeError(w, err)
		return
	}

	engine.InsertOne(order)
	// TODO:
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
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Commit()

	// TODOO:
	// Set car to idle & update the location
	res, err := tx.Exec("update cars set status='idle',location_x=?,location_y=? where id=?", order.ToX, order.ToY, order.CarID)
	rows, err := res.RowsAffected()
	if rows == 0 {

	}
	// TODO Update order's status

}
