package store

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("notification not found")

// Get returns all of a user's notifications. As a side effect, any
// notification still PENDING is promoted to DELIVERED -- fetching the list
// is what "delivery" means in this domain.
func (p *PostgreStore) Get(userId string) ([]Notification, error) {
	if _, err := p.db.Exec(
		"UPDATE notifications SET status = 'DELIVERED' WHERE user_id = $1 AND status = 'PENDING'",
		userId,
	); err != nil {
		return nil, err
	}

	rows, err := p.db.Query("SELECT id, user_id, body, status, created_at FROM notifications where user_id = $1", userId)
	if err != nil {
		return []Notification{}, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.Id, &n.UserId, &n.Body, &n.Status, &n.CreatedAt); err != nil {
			return []Notification{}, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (p *PostgreStore) Create(userId, body string) error {
	_, err := p.db.Exec("INSERT INTO notifications (user_id, body, status) VALUES ($1, $2, 'PENDING')", userId, body)
	return err
}

func (p *PostgreStore) GetStatus(id, userId string) (string, error) {
	var status string
	err := p.db.QueryRow("SELECT status FROM notifications WHERE id = $1 AND user_id = $2", id, userId).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return status, err
}

func (p *PostgreStore) UpdateStatus(id, userId, status string) (Notification, error) {
	row := p.db.QueryRow(
		`UPDATE notifications SET status = $1 WHERE id = $2 AND user_id = $3
		 RETURNING id, user_id, body, status, created_at`,
		status, id, userId,
	)
	var n Notification
	if err := row.Scan(&n.Id, &n.UserId, &n.Body, &n.Status, &n.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Notification{}, ErrNotFound
		}
		return Notification{}, fmt.Errorf("update status: %w", err)
	}
	return n, nil
}

// MarkAllRead bulk-transitions every DELIVERED notification for a user to
// READ. PENDING (not yet delivered) and DISMISSED (terminal) rows are left
// untouched -- this mirrors what a single PATCH .../status would allow one
// row at a time.
func (p *PostgreStore) MarkAllRead(userId string) error {
	_, err := p.db.Exec(
		"UPDATE notifications SET status = 'READ' WHERE user_id = $1 AND status = 'DELIVERED'",
		userId,
	)
	return err
}

func (p *PostgreStore) Delete(id, userId string) error {
	_, err := p.db.Exec("DELETE FROM notifications where id = $1 and user_id = $2", id, userId)
	return err
}
