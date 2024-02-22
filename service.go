package service

import (
	"database/sql"
	"fmt"

	"github.com/nats-io/nats.go"
)

func insertOrder(order Order) error {
	_, err := db.Exec("INSERT INTO orders (id, customer, product, quantity, total_amount) VALUES ($1, $2, $3, $4, $5)",
		order.ID, order.Customer, order.Product, order.Quantity, order.TotalAmount)
	return err
}

func getOrderFromCache(id int) (Order, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	order, ok := cache[id]
	return order, ok
}

func getOrderFromDB(id int) (Order, error) {
	var order Order
	row := db.QueryRow("SELECT * FROM orders WHERE id = $1", id)
	err := row.Scan(&order.ID, &order.Customer, &order.Product, &order.Quantity, &order.TotalAmount)
	return order, err
}

func getOrderHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	orderID := 0
	fmt.Sscanf(id, "%d", &orderID)

	// Check cache
	if order, ok := getOrderFromCache(orderID); ok {
		json.NewEncoder(w).Encode(order)
		return
	}

	// Get order from database
	order, err := getOrderFromDB(orderID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Cache order
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cache[orderID] = order

	json.NewEncoder(w).Encode(order)