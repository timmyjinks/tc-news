package database

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisStorage(host, port string) *redis.Client {
	db := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "",
		DB:       0,
	})

	return db
}
