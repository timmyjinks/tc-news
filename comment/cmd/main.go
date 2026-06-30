package main

import (
	"log"

	"github.com/timmyjinks/comment/database"
	"github.com/timmyjinks/comment/store"
)

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
