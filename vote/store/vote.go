package store

import "database/sql"

func (s *PostgreStore) Get(userId string) (int, error) {
	var sum sql.NullInt64
	err := s.db.QueryRow(
		"SELECT SUM(value) FROM votes WHERE user_id = $1", userId,
	).Scan(&sum)
	if err != nil {
		return 0, err
	}
	if !sum.Valid {
		return 0, nil
	}
	return int(sum.Int64), nil
}

func (s *PostgreStore) InsertPost(f VoteInsert) error {
	_, err := s.db.Exec("INSERT INTO votes(post_id, user_id, value) VALUES ($1, $2, $3) ON CONFLICT (user_id, post_id) DO UPDATE SET value = EXCLUDED.value", f.PostId, f.UserId, f.Value)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) InsertComment(f VoteInsert) error {
	_, err := s.db.Exec("INSERT INTO votes(comment_id, user_id, value) VALUES ($1, $2, $3) ON CONFLICT (user_id, comment_id) DO UPDATE SET value = EXCLUDED.value", f.UserId, f.CommentId, f.Value)
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
	_, err := s.db.Exec("DELETE FROM votes where comment_id = $1 and user_id = $2", commentId, userId)
	if err != nil {
		return err
	}
	return nil
}
