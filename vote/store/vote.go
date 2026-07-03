package store

func (s *PostgreStore) Get(userId string) error {
	_, err := s.db.Query("SELECT SUM(value) FROM votes where user_id = $1", userId)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) InsertPost(f VoteInsert) error {
	_, err := s.db.Exec("INSERT INTO votes(post_id, user_id, value) VALUES ($1, $2, $3 ON CONFLICT (user_id, post_id) DO UPDATE SET value = EXCLUDED.value", f.PostId, f.UserId)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) InsertComment(f VoteInsert) error {
	_, err := s.db.Exec("INSERT INTO votes(comment_id, user_id, value) VALUES ($1, $2, $3 ON CONFLICT (user_id, comment_id) DO UPDATE SET value = EXCLUDED.value", f.CommentId, f.UserId, f.Value)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) DeletePost(postId, userId string) error {
	_, err := s.db.Exec("DELETE FROM votes where post_id = $1 and user_id = $2", postId, userId)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) DeleteComment(commentId, userId string) error {
	_, err := s.db.Exec("DELETE votes where comment_id = $1 and user_id = $2", commentId, userId)
	if err != nil {
		return err
	}
	return nil
}
