package main

import (
	"log"

	"github.com/timmyjinks/notification/database"
	"github.com/timmyjinks/notification/store"
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
