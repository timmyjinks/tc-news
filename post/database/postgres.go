package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgresStorage(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS posts (
        id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        author_id  uuid NOT NULL,
				title TEXT DEFAULT '',
				body TEXT DEFAULT '{}',
				tags TEXT[] NOT NULL DEFAULT '{}',
				created_at TIMESTAMP DEFAULT now()
    )
`); err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db, nil
}
