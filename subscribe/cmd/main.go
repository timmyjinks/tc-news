package main

import (
	"log"

	"github.com/timmyjinks/follow/database"
	"github.com/timmyjinks/follow/store"
)

// @title           Follow Service API
// @version         3.0
// @description     API for subscribing to and unsubscribing from post updates.
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
		store:     store,
		jwtSecret: config.jwtSecret,
	}

	app.Run(config.addr)
}
