package store

import "database/sql"

type PostgreStore struct {
	db *sql.DB
}

func NewPostgreStore(db *sql.DB) *PostgreStore {
	return &PostgreStore{
		db: db,
	}
}
