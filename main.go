package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const apiServerAddr = "localhost:3000"
const webServerAddr = "localhost:8080"

func main() {
	holdup := make(chan bool)
	go func() {
		ApiServer()
		holdup <- true
	}()
	go func() {
		WebServer()
		holdup <- true
	}()
	<-holdup
}

func WebServer() {
	http.HandleFunc("/", webIndexHandler)
	log.Printf("starting web server at %s", webServerAddr)
	log.Fatal(http.ListenAndServe(webServerAddr, nil))
}

func webIndexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "yay!")
}

func ApiServer() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/order/{id:[0-9]+}", orderByID)
	router.HandleFunc("/api/order/{oid:[0-9]+}/lineitem/{lid:[0-9]+}", lineitemByID)
	router.HandleFunc("/api/product/{id:[0-9]+}", productByID)
	router.HandleFunc("/api/customer/{id:[0-9]+}", customerByID)
	log.Printf("starting API server at %s", apiServerAddr)
	log.Fatal(http.ListenAndServe(apiServerAddr, router))
}

type customer struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price price  `json:"price"`
}

type price int // in cents

type lineitem struct {
	ID       int     `json:"id"`
	Quantity int     `json:"quantity"`
	Product  product `json:"product"`
	SubTotal price   `json:"subtotal"`
}

type order struct {
	ID        int        `json:"id"`
	Customer  customer   `json:"customer"`
	LineItems []lineitem `json:"lineitems"`
	Total     price      `json:"total"`
	CreatedAt time.Time  `json:"created_at"`
}

var products map[int]product = map[int]product{
	1000: product{ID: 1000, Name: "Apple", Price: 250},
	1001: product{ID: 1001, Name: "Pear", Price: 325},
	1002: product{ID: 1002, Name: "Banana", Price: 175},
}

var customers map[int]customer = map[int]customer{
	111: customer{ID: 111, Name: "Fred", Email: "fred@example.com"},
	112: customer{ID: 112, Name: "Barny", Email: "barny@example.com"},
}

var orders map[int]order = map[int]order{
	1: order{
		ID:        1,
		Customer:  customers[112],
		CreatedAt: time.Date(2010, time.May, 3, 10, 30, 0, 0, time.UTC),
		LineItems: []lineitem{
			{
				ID:       0,
				Product:  products[1000],
				Quantity: 2,
				SubTotal: 2 * products[1000].Price,
			},
			{
				ID:       1,
				Product:  products[1002],
				Quantity: 1,
				SubTotal: 1 * products[1002].Price,
			},
		},
		Total: 2*products[1000].Price + 1*products[1002].Price,
	},
	2: order{
		ID:        2,
		Customer:  customers[111],
		CreatedAt: time.Date(2010, time.June, 1, 14, 5, 0, 0, time.UTC),
		LineItems: []lineitem{
			{
				ID:       0,
				Product:  products[1001],
				Quantity: 3,
				SubTotal: 3 * products[1001].Price,
			},
		},
		Total: 3 * products[1001].Price,
	},
}

func orderByID(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	order, ok := orders[orderID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(order)
}

func customerByID(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	customerID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	customer, ok := customers[customerID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(customer)
}

func productByID(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	productID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	product, ok := products[productID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(product)
}

func lineitemByID(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	orderID, err := strconv.Atoi(vars["oid"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	order, ok := orders[orderID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	lineitemID, err := strconv.Atoi(vars["lid"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var theLineItem *lineitem
	for _, li := range order.LineItems {
		if li.ID == lineitemID {
			theLineItem = &li
			break
		}
	}
	if theLineItem == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(theLineItem)
}
