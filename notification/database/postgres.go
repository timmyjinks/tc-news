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
    CREATE TABLE IF NOT EXISTS notifications (
        id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id    uuid NOT NULL,
        body       TEXT DEFAULT '',
        status     TEXT NOT NULL DEFAULT 'PENDING',
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    )
`); err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db, nil
}
