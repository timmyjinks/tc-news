package main

import (
	"log"

	"github.com/timmyjinks/user/database"
	"github.com/timmyjinks/user/store"
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
