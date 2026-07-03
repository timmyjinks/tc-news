package store

func (p *PostgreStore) Get(userId string) ([]Notification, error) {
	rows, err := p.db.Query("SELECT * FROM notifcations where user_id = $1", userId)
	if err != nil {
		return []Notification{}, err
	}
	defer rows.Close()

	var notifcations []Notification
	for rows.Next() {
		var notification Notification
		err := rows.Scan(&notification.Id, &notification.UserId, &notification.Read, &notification.CreatedAt)
		if err != nil {
			return []Notification{}, err
		}
		notifcations = append(notifcations, notification)
	}

	return notifcations, nil
}

func (p *PostgreStore) Update(id, userId string) error {
	_, err := p.db.Exec("UPDATE notifcations SET read = true where id = $1 and user_id = $2", id, userId)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgreStore) UpdateAll(userId string) error {
	_, err := p.db.Exec("UPDATE notifcations SET read = true where user_id = $1", userId)
	if err != nil {
		return err
	}
	return nil
}
