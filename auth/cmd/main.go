package main

import (
	"github.com/timmyjinks/auth/database"
	"github.com/timmyjinks/auth/store"
)

// @title           Auth Service API
// @version         3.0
// @description     Handles login and token refresh for the auth service.
// @host            localhost:8080
// @BasePath        /

func main() {
	config := Load()

	db := database.NewRedisStorage()
	store := store.NewRedisStore(db)

	app := application{
		store:     store,
		jwtSecret: "testing",
	}

	app.Run(config.addr)
}
