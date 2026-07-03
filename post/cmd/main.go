package main

import (
	"log"

	"github.com/timmyjinks/post/database"
	"github.com/timmyjinks/post/store"
)

// @title           Post Service API
// @version         3.0
// @description     Handles creating, reading, updating, and deleting posts.
// @host            localhost:8081
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
