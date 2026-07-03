package store

import "time"

type Comment struct {
	Id        string    `sql:"id"`
	PostId    string    `sql:"post_id"`
	UserId    string    `sql:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentCreate struct {
	ParentId string `json:"parent_id"`
	PostId   string `json:"post_id"`
	UserId   string `json:"user_id"`
	Body     string `json:"body"`
}

type CommentUpdate struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`
	Body   string `json:"body"`
}
