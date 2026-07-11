package main

import (
	"log"

	"github.com/timmyjinks/vote/database"
	"github.com/timmyjinks/vote/store"
)

// @title           Vote Service API
// @version         3.0
// @description     API for casting and removing votes on posts and comments.
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
