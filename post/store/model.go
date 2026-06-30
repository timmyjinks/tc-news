package store

import "time"

type Post struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type PostCreate struct {
	UserId string
	Body   string
}

type PostUpdate struct {
	PostId string
	UserId string
	Body   string
}
