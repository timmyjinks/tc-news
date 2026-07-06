package main

import (
	"log"
	"net/http"
	"time"

	"github.com/timmyjinks/comment/database"
	"github.com/timmyjinks/comment/kafka"
	"github.com/timmyjinks/comment/store"
)

// @title           Comment Service API
// @version         3.0
// @description     API for creating, reading, updating, and deleting comments on posts.
// @host            localhost:8080
// @BasePath        /

func main() {
	config := Load()

	db, err := database.NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewPostgreStore(db)
	queue := kafka.NewKafkaService("notifications")

	app := application{
		store:           store,
		producer:        queue,
		postServiceAddr: config.postServiceAddr,
		httpClient:      &http.Client{Timeout: 5 * time.Second},
	}

	app.Run(config.addr)
}
