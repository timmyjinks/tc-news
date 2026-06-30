package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgresStorage() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://postgres:password@user-db:5432/postgres?sslmode=disable")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS users (
        id         uuid PRIMARY KEY default gen_random_uuid(),
        name TEXT NOT NULL,
				created_at TIMESTAMP Default now()
    )
`); err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db, nil
}
