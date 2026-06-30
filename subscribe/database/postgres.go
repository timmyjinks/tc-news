package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgresStorage() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://postgres:password@subscribe-db:5432/postgres?sslmode=disable")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS subscribers (
        id         uuid PRIMARY KEY default gen_random_uuid(),
        post_id    uuid NOT NULL, 
        user_id    uuid NOT NULL,
        UNIQUE(post_id, user_id)
    )
`); err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db, nil
}
