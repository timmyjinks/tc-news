package store

func (s *PostgreStore) GetById(commentId string) (Comment, error) {
	row := s.db.QueryRow("SELECT * from comments where id = $1", commentId)

	var comment Comment
	err := row.Scan(&comment.Id, &comment.PostId, &comment.UserId, &comment.Body, &comment.CreatedAt)
	if err != nil {
		return Comment{}, err
	}

	return comment, nil
}

func (s *PostgreStore) Get(postId string) ([]Comment, error) {
	rows, err := s.db.Query("SELECT * from comments where post_id = $1", postId)
	if err != nil {
		return []Comment{}, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.Id, &comment.PostId, &comment.UserId, &comment.Body, &comment.CreatedAt)
		if err != nil {
			return []Comment{}, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (s *PostgreStore) Create(f CommentCreate) error {
	_, err := s.db.Exec("INSERT INTO comments (post_id, user_id, body) VALUES ($1, $2, $3)", f.PostId, f.UserId, f.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) Update(f CommentUpdate) error {
	s.db.Exec("UPDATE comments SET body = $1 where id = $2 and user_id = $3", f.Id, f.UserId)
	return nil
}

func (s *PostgreStore) Delete(id, userId string) error {
	s.db.Exec("DELETE comments where id = $1 and user_id = $2", id, userId)
	return nil
}
