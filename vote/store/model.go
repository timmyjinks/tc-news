package store

type Vote struct {
	Id     string `sql:"id"`
	PostId string `sql:"post_id"`
	UserId string `sql:"user_id"`
	Value  int    `json:"value"`
}

type VoteInsert struct {
	Id        string `json:"id"`
	PostId    string `json:"post_id"`
	CommentId string `json:"comment_id"`
	UserId    string `json:"user_id"`
	Value     int    `json:"value"`
}
