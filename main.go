package main

import (
	"database/sql"
	"encoding/json"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	_ "github.com/lib/pq"
)


const (
	postgresHost     = "localhost"
	postgresPort     = "5432"
	postgresUser     = "your_username"
	postgresPassword = "your_password"
	postgresDBName   = "your_database_name"

	natsURL      = "nats://localhost:4222"
	clusterID    = "test-cluster"
	clientID     = "test-client"
	subject      = "orders"
	cacheSize    = 1000
	cacheTimeout = 60 // seconds
)

type Order struct {
	ID          int    `json:"id"`
	Customer    string `json:"customer"`
	Product     string `json:"product"`
	Quantity    int    `json:"quantity"`
	TotalAmount int    `json:"total_amount"`
}

var (
	db         *sql.DB
	cache      = make(map[int]Order)
	cacheMutex sync.RWMutex
)

func check(ctx context.Context, msg <-chan[]byte, conn *pgx.Conn) {
	logger := log.LoggerFromContext(ctx)
	logger.Info("checking messages")
}

go func() {
    for {
      select {
      case msg := <-msg:
        msgStruct := &validator.Order{}
        if err := json.Unmarshal(msg, msgStruct); err != nil {
          logger.Error("JSON bad data:", zap.Error(err))
          break
        }

        if err := msgStruct.Validate(ctx); err != nil {
          logger.Error("Validation error:", zap.Error(err))
          break
        }

        if err := postgres.Create(ctx, conn, msgStruct); err != nil {
          logger.Error("postgres.Create() error")
          break
        }
        cache.CACHE[msgStruct.OrderUID] = msgStruct
      case <-ctx.Done():
        logger.Info("Stop listening messages...")
        return
      }
    }
  }()

func main() {
	// Connect to PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDBName)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Connect to NATS Streaming
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Error connecting to NATS:", err)
	}
	defer nc.Close()

	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Fatal("Error connecting to NATS Streaming:", err)
	}
	defer sc.Close()

	// Subscribe to NATS Streaming subject
	_, err = sc.Subscribe(subject, func(msg *stan.Msg) {
		var order Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Println("Error unmarshalling JSON:", err)
			return
		}

		// Insert order into PostgreSQL
		if err := insertOrder(order); err != nil {
			log.Println("Error inserting order into PostgreSQL:", err)
			return
		}

		// Cache order
		cacheMutex.Lock()
		defer cacheMutex.Unlock()
		cache[order.ID] = order
	}, stan.DurableName("my-durable"))
	if err != nil {
		log.Fatal("Error subscribing to NATS Streaming subject:", err)
	}

	// HTTP handler
	http.HandleFunc("/order", getOrderHandler)

	// Start HTTP server
	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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
}
