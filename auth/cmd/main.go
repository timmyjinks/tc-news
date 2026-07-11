package main

import (
	"log"

	"github.com/timmyjinks/auth/database"
	"github.com/timmyjinks/auth/store"
)

// @title           Auth Service API
// @version         3.0
// @description     Handles login, token refresh, and user management.
// @host            localhost:8085
// @BasePath        /

func main() {
	config := Load()

	redisDb := database.NewRedisStorage(config.sessionDBHost, config.sessionDBPort)
	redisStore := store.NewRedisStore(redisDb)

	postgresDb, err := database.NewPostgresStorage(config.userDBHost, config.userDBPort)
	if err != nil {
		log.Fatal(err)
	}
	postgresStore := store.NewPostgresStore(postgresDb)

	app := application{
		store:     redisStore,
		userStore: postgresStore,
		jwtSecret: "testing",
	}

	app.Run(config.addr)
}
