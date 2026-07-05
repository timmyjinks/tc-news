package store

import (
	"fmt"

	"github.com/lib/pq"
)

const postColumns = "id, author_id, title, body, tags, created_at"

func (s *PostgreStore) GetById(id string) (Post, error) {
	row := s.db.QueryRow("SELECT "+postColumns+" from posts where id = $1", id)

	var post Post
	err := row.Scan(&post.Id, &post.AuthorId, &post.Title, &post.Body, pq.Array(&post.Tags), &post.CreatedAt)
	if err != nil {
		return Post{}, err
	}
	return post, nil
}

func (s *PostgreStore) Get(params ListPostsParams) ([]Post, error) {
	query := "SELECT " + postColumns + " FROM posts"
	args := []interface{}{}

	if params.Tag != "" {
		args = append(args, params.Tag)
		query += fmt.Sprintf(" WHERE $%d = ANY(tags)", len(args))
	}

	order := "DESC"
	if params.Sort == "oldest" {
		order = "ASC"
	}
	query += " ORDER BY created_at " + order

	if params.Limit > 0 {
		args = append(args, params.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if params.Offset > 0 {
		args = append(args, params.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []Post{}, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.AuthorId, &post.Title, &post.Body, pq.Array(&post.Tags), &post.CreatedAt)
		if err != nil {
			return []Post{}, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *PostgreStore) Create(f PostCreate) error {
	_, err := s.db.Exec(
		"INSERT INTO posts (author_id, title, body, tags) VALUES ($1, $2, $3, $4)",
		f.AuthorId, f.Title, f.Body, pq.Array(f.Tags),
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) Update(f PostUpdate) error {
	_, err := s.db.Exec(
		"UPDATE posts SET title = $1, body = $2, tags = $3 where id = $4 and author_id = $5",
		f.Title, f.Body, pq.Array(f.Tags), f.PostId, f.AuthorId,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) Delete(postId, authorId string) error {
	_, err := s.db.Exec("DELETE from posts where id = $1 and author_id = $2", postId, authorId)
	if err != nil {
		return err
	}
	return nil
}
