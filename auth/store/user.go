package store

func (s *PostgresStore) Get() ([]User, error) {
	rows, err := s.db.Query("SELECT id, name, password_hash, created_at FROM users")
	if err != nil {
		return []User{}, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Id, &user.Name, &user.PasswordHash, &user.CreatedAt); err != nil {
			return []User{}, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *PostgresStore) GetById(id string) (User, error) {
	row := s.db.QueryRow("SELECT id, name, password_hash, created_at FROM users WHERE id = $1", id)

	var user User
	if err := row.Scan(&user.Id, &user.Name, &user.PasswordHash, &user.CreatedAt); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *PostgresStore) GetByName(name string) (User, error) {
	row := s.db.QueryRow("SELECT id, name, password_hash, created_at FROM users WHERE name = $1", name)

	var user User
	if err := row.Scan(&user.Id, &user.Name, &user.PasswordHash, &user.CreatedAt); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *PostgresStore) Create(f UserCreate) error {
	_, err := s.db.Exec("INSERT INTO users (name, password_hash) VALUES ($1, $2)", f.Name, f.PasswordHash)
	return err
}

func (s *PostgresStore) Update(f UserUpdate) error {
	_, err := s.db.Exec("UPDATE users SET name = $1 where id = $2", f.Name, f.Id)
	if err != nil {
		return err
	}
	return err
}

func (s *PostgresStore) Delete(userId string) error {
	_, err := s.db.Exec("DELETE FROM users where id = $1", userId)
	return err
}
