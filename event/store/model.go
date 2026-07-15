package store

import "time"

type Event struct {
	Id        string    `sql:"id"`
	Type      string    `sql:"user_id"`
	Body      string    `sql:"body"`
	CreatedAt time.Time `sql:"created_at"`
}
