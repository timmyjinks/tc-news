package store

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	db *redis.Client
}

type PostgresStore struct {
	db *sql.DB
}

func NewRedisStore(db *redis.Client) *RedisStore {
	return &RedisStore{
		db: db,
	}
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{
		db: db,
	}
}
