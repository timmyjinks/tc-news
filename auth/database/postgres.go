package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgresStorage(host, port string) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://postgres:password@%s:%s/postgres?sslmode=disable", host, port)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS users (
        id            uuid PRIMARY KEY default gen_random_uuid(),
        name          TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        created_at    TIMESTAMP DEFAULT now()
    )
`); err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db, nil
}
