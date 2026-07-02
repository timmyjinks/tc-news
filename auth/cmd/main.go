package main

import (
	"github.com/timmyjinks/auth/database"
	"github.com/timmyjinks/auth/store"
)

func main() {
	config := Load()

	db := database.NewRedisStorage()
	store := store.NewRedisStore(db)

	app := application{
		store: store,
	}

	app.Run(config.addr)
}
