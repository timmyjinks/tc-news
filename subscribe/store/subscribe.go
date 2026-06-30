package store

func (s *PostgreStore) Get(userId string) ([]Subscriber, error) {
	rows, err := s.db.Query("SELECT * from subscribers where user_id = $1", userId)
	if err != nil {
		return []Subscriber{}, err
	}
	defer rows.Close()

	var follows []Subscriber
	for rows.Next() {
		var follow Subscriber
		err := rows.Scan(&follow)
		if err != nil {
			return []Subscriber{}, err
		}
		follows = append(follows, follow)
	}

	return follows, nil
}

func (s *PostgreStore) Create(f SubscriberCreate) error {
	_, err := s.db.Exec("INSERT INTO subscribers (post_id, user_id) VALUES ($1, $2)", f.PostId, f.UserId)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) Delete(postId, userId string) error {
	_, err := s.db.Exec("DELETE subscribers where post_id = $1 and user_id = $2", postId, userId)
	if err != nil {
		return err
	}
	return nil
}
