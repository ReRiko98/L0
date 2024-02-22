package main

import (
	"database/sql"
	"fmt"
	"log"
	"myapp/internal/manager"
	"myapp/internal/service"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

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