package store

func (p *PostgreStore) Get(userId string) ([]Notification, error) {
	rows, err := p.db.Query("SELECT * FROM notifications where user_id = $1", userId)
	if err != nil {
		return []Notification{}, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notification Notification
		err := rows.Scan(&notification.Id, &notification.UserId, &notification.Body, &notification.Read, &notification.CreatedAt)
		if err != nil {
			return []Notification{}, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (p *PostgreStore) Create(userId, body string) error {
	_, err := p.db.Exec("INSERT INTO notifications (user_id, body) VALUES ($1, $2)", userId, body)
	return err
}

func (p *PostgreStore) Update(id, userId string) error {
	_, err := p.db.Exec("UPDATE notifications SET read = true where id = $1 and user_id = $2", id, userId)
	return err
}

func (p *PostgreStore) UpdateAll(userId string) error {
	_, err := p.db.Exec("UPDATE notifications SET read = true where user_id = $1", userId)
	return err
}

func (p *PostgreStore) Delete(id, userId string) error {
	_, err := p.db.Exec("DELETE FROM notifications where id = $1 and user_id = $2", id, userId)
	return err
}
