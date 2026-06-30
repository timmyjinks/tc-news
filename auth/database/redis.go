package database

import (
	"github.com/redis/go-redis/v9"
)

func NewRedisStorage() *redis.Client {
	db := redis.NewClient(&redis.Options{
		Addr:     "auth-db:6379",
		Password: "",
		DB:       0,
	})

	return db
}
