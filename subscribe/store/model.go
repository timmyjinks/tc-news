package store

type Subscriber struct {
	Id     string `sql:"id"`
	PostId string `sql:"post_id"`
	UserId string `sql:"user_id"`
}

type SubscriberCreate struct {
	PostId string `json:"post_id"`
	UserId string `json:"user_id"`
}

type SubscriberUpdate struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`
}
