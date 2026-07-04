package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgresStorage() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://postgres:password@comment-db:5432/postgres?sslmode=disable")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS comments (
        id         uuid PRIMARY KEY default gen_random_uuid(),
				parent_id  uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid, 
        post_id    uuid NOT NULL, 
        user_id    uuid NOT NULL,
				body TEXT DEFAULT '{}',
				created_at TIMESTAMP DEFAULT now()
    )
`); err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db, nil
}
