package service

import (
	"fmt"
	"sync"
	"context"

	"github.com/nats-io/nats.go"
	"github.com/jackc/pgx/v4"
)

type Cache struct {
	OrderUID string `json:"order_uid"`
}

type DatabaseService struct {
	db    *pgx.Conn
	nc    *nats.Conn
	cache map[int]Cache
	mu    sync.Mutex // save cache
}

func NewDS(db *pgx.Conn, nc *nats.Conn) *DatabaseService {
	return &DatabaseService{
		db:    db,
		nc:    nc,
		cache: make(map[int]Cache),
	}
}

func (s *DatabaseService) GetInfo(number int) (string, error) {
	var jsonData string
	err := s.db.QueryRow(context.Background(), "SELECT Name_Json_Info FROM Json_Info WHERE ID_Json_Info = $1", number).Scan(&jsonData)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	message := fmt.Sprintf("call complete: %d", number)
	s.nc.Publish("log", []byte(message))

	return jsonData, nil
}

func (s *DatabaseService) AddData(jsonData string) (int, error) {
	var newID int
	_, err := s.db.Exec(context.Background(), "INSERT INTO Json_Info (Name_Json_Info) VALUES ($1)", jsonData)
	if err != nil {
		return 0, fmt.Errorf("adding error: %v", err)
	}

	// last modificator
	err = s.db.QueryRow(context.Background(), "SELECT lastval()").Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("error get Info: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	newCacheData := Cache{OrderUID: jsonData}
	s.cache[newID] = newCacheData

	message := fmt.Sprintf("New data added: %s", jsonData)
	s.nc.Publish("log", []byte(message))

	return newID, nil
}
