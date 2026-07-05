package store

import "time"

type Post struct {
	Id        string    `json:"id"`
	AuthorId  string    `json:"author_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type PostCreate struct {
	AuthorId string
	Title    string
	Body     string
	Tags     []string
}

type PostUpdate struct {
	PostId   string
	AuthorId string
	Title    string
	Body     string
	Tags     []string
}

type ListPostsParams struct {
	Limit  int
	Offset int
	Tag    string
	Sort   string
}
