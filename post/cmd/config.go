package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	addr      string
	dbURI     string
	kafkaAddr string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
	}

	return Config{
		addr:      getEnv("ADDR", ":8080"),
		dbURI:     buildDBURI(),
		kafkaAddr: getEnv("KAFKA_ADDR", "kafka-service:9092"),
	}
}

func buildDBURI() string {
	if uri := os.Getenv("DATABASE_URL"); uri != "" {
		return uri
	}

	host := getEnv("DB_HOST", "post-db")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "postgres")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslmode)
}

func getEnv(key, fallback string) string {
	env := os.Getenv(key)
	if env == "" {
		return fallback
	}
	return env
}
