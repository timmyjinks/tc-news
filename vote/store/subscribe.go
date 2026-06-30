package store

func (s *PostgreStore) Create(f VoteCreate) error {
	s.db.Exec("INSERT INTO votes (post_id, user_id) VALUES ($1, $2)", f.PostId, f.UserId)
	return nil
}

func (s *PostgreStore) Update(f VoteUpdate) error {
	s.db.Exec("INSERT INTO subsribers (post_id, user_id) VALUES ($1, $2)", f.PostId, f.UserId)
	return nil
}

func (s *PostgreStore) Delete(postId, userId string) error {
	s.db.Exec("DELETE subsribers where post_id = $1 and user_id = $2", postId, userId)
	return nil
}
