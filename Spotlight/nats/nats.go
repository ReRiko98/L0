package main

import (
    "fmt"
    "log"
    "os"

    "github.com/nats-io/nats.go"
)

func main() {
    natsURL := "localhost" 
    natsPort := os.Getenv("NATS_PORT")
    if natsPort == "" {
        natsPort = "4222"
    }
    natsAddress := fmt.Sprintf("%s:%s", natsURL, natsPort)

    log.Printf("Nats connect: %s\n", natsAddress)

    // NATS
    nc, err := nats.Connect(natsAddress)
    if err != nil {
        log.Fatalf("Nats connect error: %v\n", err)
    }
    defer nc.Close()

    //  log
    _, err = nc.Subscribe("log", func(m *nats.Msg) {
        log.Printf("Message recieved: %s\n", string(m.Data))
    })
    if err != nil {
        log.Fatalf("Error: %v\n", err)
    }

    // Program will continue work
    select {}
}