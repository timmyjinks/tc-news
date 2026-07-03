package store

import "time"

type Notification struct {
	Id        string    `sql:"id"`
	UserId    string    `sql:"user_id"`
	Body      string    `sql:"body"`
	Read      bool      `sql:"read"`
	CreatedAt time.Time `sql:"created_at"`
}

type NotificationUpdate struct {
	Id        string
	UserId    string
	Body      string
	Read      bool
	CreatedAt time.Time
}
