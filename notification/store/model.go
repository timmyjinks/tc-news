package store

import "time"

type Notification struct {
	Id        string    `sql:"id"`
	UserId    string    `sql:"user_id"`
	Body      string    `sql:"body"`
	Status    string    `sql:"status"`
	CreatedAt time.Time `sql:"created_at"`
}

type NotificationUpdate struct {
	Id        string
	UserId    string
	Body      string
	Status    string
	CreatedAt time.Time
}
