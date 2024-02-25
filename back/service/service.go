package service

import (
	"database/sql"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Cache struct {
	OrderUID string `json:"order_uid"`
}

type DS struct {
	db    *sql.DB
	nc    *nats.Conn
	cache map[int]Cache
}

func NewDS(db *sql.DB, nc *nats.Conn) *DS {
	return &DS{
		db:    db,
		nc:    nc,
		cache: make(map[int]Cache),
	}
}

func (s *DS) GetInfo(number int) (string, error) {
	var jsonData string
	//Get from BD
	err := s.db.QueryRow("SELECT Name_Json_Info FROM Json_Info WHERE ID_Json_Info = $1", number).Scan(&jsonData)
	if err != nil {
		return "", fmt.Errorf("Call error")
	}

	// Send a message to Nuts
	message := fmt.Sprintf("Call complete id: %d", number)
	s.nc.Publish("log", []byte(message))

	return jsonData, nil
}

func (s *DS) AddData(jsonData string) (int, error) {
	// New Data in Cache
	var newID int
	err := s.db.QueryRow("INSERT INTO Json_Info (Name_Json_Info) VALUES ($1) RETURNING ID_Json_Info", jsonData).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при добавлении данных в базу: %v", err)
	}

	// Save data in Cache
	newCacheData := Cache{OrderUID: jsonData}
	s.cache[newID] = newCacheData

	// Send message to Nuts
	message := fmt.Sprintf("Добавлены новые данные: %s", jsonData)
	s.nc.Publish("log", []byte(message))

	return newID, nil
}