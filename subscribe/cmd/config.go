package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	addr      string
	grpcAddr  string
	jwtSecret string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
	}

	return Config{
		addr:      getEnv("ADDR", ":8080"),
		grpcAddr:  getEnv("GRPC_ADDR", ":9090"),
		jwtSecret: getEnv("JWT_SECRET", "testing"),
	}
}

func getEnv(key, fallback string) string {
	env := os.Getenv(key)
	if env == "" {
		return fallback
	}
	return env
}
