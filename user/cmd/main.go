package main

import (
	"log"

	"github.com/timmyjinks/user/database"
	"github.com/timmyjinks/user/store"
)

// @title           User Service API
// @version         3.0
// @description     API for creating, reading, updating, and deleting users.
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

	app.Run(config.addr)
}
