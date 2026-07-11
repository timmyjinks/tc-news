package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	addr          string
	userDBHost    string
	userDBPort    string
	sessionDBHost string
	sessionDBPort string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
	}

	return Config{
		addr:          getEnv("ADDR", ":8080"),
		userDBHost:    getEnv("USER_DB_HOST", "auth-postgres-db"),
		userDBPort:    getEnv("USER_DB_PORT", "5432"),
		sessionDBHost: getEnv("SESSION_DB_HOST", "auth-db"),
		sessionDBPort: getEnv("SESSION_DB_PORT", "6379"),
	}
}

func getEnv(key, fallback string) string {
	env := os.Getenv(key)
	if env == "" {
		return fallback
	}
	return env
}
