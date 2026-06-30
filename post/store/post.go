package store

func (s *PostgreStore) Get(userId string) ([]Post, error) {
	rows, err := s.db.Query("SELECT * from posts where user_id = $1", userId)
	if err != nil {
		return []Post{}, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.UserId, &post.Body, &post.CreatedAt)
		if err != nil {
			return []Post{}, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *PostgreStore) Create(f PostCreate) error {
	_, err := s.db.Exec("INSERT INTO posts (user_id, body) VALUES ($1, $2)", f.UserId, f.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) Update(f PostUpdate) error {
	_, err := s.db.Exec("Update posts SET body = $1 where id = $2 and user_id = $3", f.Body, f.PostId, f.UserId)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) Delete(postId, userId string) error {
	s.db.Exec("DELETE posts where id = $1 and user_id = $2", postId, userId)
	return nil
}
