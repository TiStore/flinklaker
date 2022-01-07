package main

import (
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
)

var engine *xorm.Engine

func main() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", YourHandler)

	r.HandleFunc("/car/{id}", CarHandler)
	r.HandleFunc("/order/{id}", OrderHandler)
	r.HandleFunc("/order", OrderHandler)
	var err error
	engine, err = xorm.NewEngine("mysql", "root:@tcp(127.0.0.1:4000)/test?charset=utf8")
	if err != nil {
		panic(err)
	}
	err = engine.Sync2(new(Car))
	if err != nil {
		panic(err)
	}
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":7998", r))
}

func YourHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello TiLaker!\n"))
}

func CarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		PutNewCar(w, r)
	} else if r.Method == http.MethodDelete {
		DeleteCar(w, r)
	}
}

func OrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		PutNewOrder(w, r)
	} else if r.Method == http.MethodDelete {
		DeleteOrder(w, r)
	}
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write([]byte(err.Error()))
}
