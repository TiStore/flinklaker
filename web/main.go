package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
)

var engine *xorm.Engine

var ra *rand.Rand

var sleepCheck int = 2
var lock sync.Mutex

func main() {
	ra = rand.New(rand.NewSource(time.Now().Unix()))
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", YourHandler)

	r.HandleFunc("/car/{id}", CarHandler)
	r.HandleFunc("/order/{id}", OrderHandler)
	r.HandleFunc("/orderDelay/{id}", OrderDelayHandler)
	r.HandleFunc("/sleep", SetSleepHandler)
	r.HandleFunc("/order", OrderHandler)
	r.HandleFunc("/location", LocationHandler)
	var err error
	engine, err = xorm.NewEngine("mysql", "root:@tcp(127.0.0.1:4000)/test?charset=utf8")
	if err != nil {
		panic(err)
	}
	err = engine.Sync2(new(Car))
	if err != nil {
		panic(err)
	}
	go CheckAndRunninigOrders()
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":7998", r))
}

func CheckAndRunninigOrders() {
	for {
		lock.Lock()
		time.Sleep(time.Duration(sleepCheck) * time.Second) //just demo
		err := checkAndRunOrders()
		log.Println("Finished check and process table nearcars", err)
		lock.Unlock()
	}
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

func SetSleepHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		t, err := strconv.ParseInt(r.URL.Query().Get("time"), 10, 64)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		lock.Lock()
		sleepCheck = int(t)
		lock.Unlock()
		w.Write([]byte("ok"))
	} else {
		w.Write([]byte(fmt.Sprintf("%d", sleepCheck)))
	}

}

func OrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		PutNewOrder(w, r)
	} else if r.Method == http.MethodDelete {
		DeleteOrder(w, r)
	}
}

func OrderDelayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		PutNewDelayOrder(w, r)
	} else if r.Method == http.MethodDelete {
		DeleteOrder(w, r)
	}
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write([]byte(err.Error()))
}

func writeJson(w http.ResponseWriter, data interface{}) {
	jdata, _ := json.Marshal(data)
	w.Write(jdata)
}

type Location struct {
	Id int64
	X  float64
	Y  float64
}

// GET /location
// get a location randomly
// Example: curl -X GET "http://localhost:8000/location"
// {"Id":20424,"X":40.25,"Y":115.53}
func LocationHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("select id,x,y from locations where id=%d\n", ra.Int63n(50)+50002)
	raws := engine.DB().QueryRow("select id,x,y from locations where id=?", ra.Int63n(50)+50002)
	var loca Location
	err := raws.Scan(&loca.Id, &loca.X, &loca.Y)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJson(w, loca)
}
