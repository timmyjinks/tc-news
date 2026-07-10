package main

import (
	"context"
	"log"

	"github.com/timmyjinks/notification/database"
	"github.com/timmyjinks/notification/httpclient"
	"github.com/timmyjinks/notification/kafka"
	"github.com/timmyjinks/notification/store"
)

// @title           Notification Service API
// @version         3.0
// @description     API for listing and marking user notifications as read.
// @host            localhost:8080
// @BasePath        /

func main() {
	config := Load()

	db, err := database.NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewPostgreStore(db)

	app := application{
		store: store,
	}

	subscribeClient := httpclient.NewSubscribeClient(config.subscribeHTTPAddr)

	queue := kafka.NewKafkaService("notifications")

	ctx := context.Background()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[INFO] notification worker shutting down")
				return
			default:
				msg, err := queue.Consumer.Read(context.Background())
				if err != nil {
					log.Println("[WARN]", err)
					continue
				}
				if err := handleMessage(store, subscribeClient, msg); err != nil {
					log.Println("[WARN]", err)
				}
			}
		}
	}()

	app.Run(config.addr)

}
