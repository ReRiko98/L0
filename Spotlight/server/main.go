package main

import (
	"database/sql"
	"fmt"
	"log"
	"myapp/back/handler"
	"myapp/back/service"
	"net/http"

	_"github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

var (
	db *sql.DB
	nt *nats.Conn
)

func main() {
	//DB Connect
	connStr := "user=postgres password=postgres dbname=l0-postgres-1 host=localhost port=5432 sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//NATS Connect
	natsURL := "localhost"
	log.Println("Connect to Nats...")
	nt, err = nats.Connect(natsURL)
	if err != nil {
		log.Println("Nats error connection")
		log.Fatal(err)
	}
	log.Println("Nats connect successful")
	defer nt.Close()

	databaseService := service.NewDS(db, nt)
	databaseHandler := handler.NewDatabaseCheck(databaseService)

	http.HandleFunc("/", databaseHandler.Index)
	http.HandleFunc("/get-info", databaseHandler.GetInfo)
	http.HandleFunc("/add-info", databaseHandler.AddData)

	fmt.Println("Load server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}