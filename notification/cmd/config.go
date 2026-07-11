package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	addr              string
	subscribeGRPCAddr string
	subscribeHTTPAddr string
	jwtSecret         string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
	}

	return Config{
		addr:              getEnv("ADDR", ":8080"),
		subscribeGRPCAddr: getEnv("SUBSCRIBE_GRPC_ADDR", "subscription:9090"),
		subscribeHTTPAddr: getEnv("SUBSCRIBE_HTTP_ADDR", "http://subscription:8080"),
		jwtSecret:         getEnv("JWT_SECRET", "testing"),
	}
}

func getEnv(key, fallback string) string {
	env := os.Getenv(key)
	if env == "" {
		return fallback
	}
	return env
}
