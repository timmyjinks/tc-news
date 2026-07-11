package store

import "time"

type Data struct {
	Id   string        `redis:"id"`
	Name string        `redis:"name"`
	TTL  time.Duration `redis:"ttl"`
}

type User struct {
	Id           string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
}

type UserCreate struct {
	Name         string
	PasswordHash string
}

type UserUpdate struct {
	Id   string
	Name string
}
