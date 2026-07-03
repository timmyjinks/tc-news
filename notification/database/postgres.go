package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgresStorage() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://postgres:password@notification-db:5432/postgres?sslmode=disable")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS notifications(
        id         SERIAL PRIMARY KEY,
        user_id    INT NOT NULL,
				body 			 TEXT, 
				read       BOOLEAN NOT NULL DEFAULT false
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    )
`); err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db, nil
}
