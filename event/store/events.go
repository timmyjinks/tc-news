package store

func (p *PostgreStore) Get() ([]Event, error) {

	rows, err := p.db.Query("SELECT * FROM events")
	if err != nil {
		return []Event{}, err
	}
	defer rows.Close()

	var notifications []Event
	for rows.Next() {
		var n Event
		if err := rows.Scan(&n.Id, &n.Type, &n.Body, &n.CreatedAt); err != nil {
			return []Event{}, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (p *PostgreStore) Create(_type, body string) error {
	_, err := p.db.Exec("INSERT INTO events (type, body) VALUES ($1, $2)", _type, body)
	return err
}
