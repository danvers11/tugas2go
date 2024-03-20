package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Structs untuk menyimpan data
type Item struct {
	ItemCode    string `json:"itemCode"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

type Order struct {
	OrderID      int    `json:"orderID"`
	CustomerName string `json:"customerName"`
	OrderedAt    string `json:"orderedAt"`
	Items        []Item `json:"items"`
}

var orders []Order
var orderIDCounter int

// Fungsi handler untuk membuat pesanan baru
func createOrder(w http.ResponseWriter, r *http.Request) {
	var newOrder Order
	json.NewDecoder(r.Body).Decode(&newOrder)
	newOrder.OrderID = orderIDCounter
	orderIDCounter++

	orders = append(orders, newOrder)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newOrder)
}

// Fungsi handler untuk mendapatkan semua pesanan
func getOrders(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(orders)
}

// Fungsi handler untuk mengupdate pesanan
func updateOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	orderId, err := strconv.Atoi(params["orderId"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var updatedOrder Order
	json.NewDecoder(r.Body).Decode(&updatedOrder)

	for i, order := range orders {
		if order.OrderID == orderId {
			// Update customerName dan orderedAt
			orders[i].CustomerName = updatedOrder.CustomerName
			orders[i].OrderedAt = updatedOrder.OrderedAt

			// Jika ada item yang dimasukkan, kita juga bisa memperbarui item
			if len(updatedOrder.Items) > 0 {
				orders[i].Items = updatedOrder.Items
			}

			json.NewEncoder(w).Encode(orders[i])
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

// Fungsi handler untuk menghapus pesanan
func deleteOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	orderId, err := strconv.Atoi(params["orderId"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for i, order := range orders {
		if order.OrderID == orderId {
			orders = append(orders[:i], orders[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	router := mux.NewRouter()

	// Menginisialisasi beberapa pesanan untuk contoh
	orders = append(orders, Order{
		OrderID:      1,
		CustomerName: "Tom Jerry",
		OrderedAt:    "2019-11-09T21:21:46+00:00",
		Items: []Item{
			{
				ItemCode:    "123",
				Description: "IPhone 10x",
				Quantity:    1,
			},
		},
	})

	orderIDCounter = 2

	// Menambahkan endpoint-endpoint ke router
	router.HandleFunc("/orders", createOrder).Methods("POST")
	router.HandleFunc("/orders", getOrders).Methods("GET")
	router.HandleFunc("/orders/{orderId}", updateOrder).Methods("PUT")
	router.HandleFunc("/orders/{orderId}", deleteOrder).Methods("DELETE")

	// Menjalankan server
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
