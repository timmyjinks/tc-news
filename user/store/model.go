package store

import "time"

type User struct {
	Id        string    `sql:"id"`
	Name      string    `sql:"name"`
	CreatedAt time.Time `sql:"createAt"`
}

type UserCreate struct {
	Name string
}

type UserUpdate struct {
	Id   string
	Name string
}
