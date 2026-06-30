package store

func (s *PostgreStore) Get() ([]User, error) {
	rows, err := s.db.Query("SELECT * FROM users")
	if err != nil {
		return []User{}, err
	}

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user)
		if err != nil {
			return []User{}, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *PostgreStore) GetById(id string) (User, error) {
	row := s.db.QueryRow("SELECT * FROM users where id = $1", id)

	var user User
	err := row.Scan(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *PostgreStore) Create(f UserCreate) error {
	_, err := s.db.Exec("INSERT INTO users (name) VALUES ($1)", f.Name)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgreStore) Update(f UserUpdate) error {
	s.db.Exec("UPDATE users set name = $1 where user_id = $2", f.Name, f.Id)
	return nil
}

func (s *PostgreStore) Delete(userId string) error {
	s.db.Exec("DELETE users where user_id = $1", userId)
	return nil
}
