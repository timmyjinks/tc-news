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
    CREATE TABLE IF NOT EXISTS votes (
        id         SERIAL PRIMARY KEY,
        post_id    uuid, 
        comment_id uuid,
        user_id    uuid NOT NULL,
				value 		 INT NOT NULl,
 				CONSTRAINT uniq_user_post UNIQUE (user_id, post_id),
				CONSTRAINT uniq_user_comment UNIQUE (user_id, comment_id)
    )
`); err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db, nil
}
