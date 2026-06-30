package store

func (s *PostgresStore) GetByNameAndPassword(name, password string) error {
	row := s.db.QueryRow("SELECT * FROM users where name = $1 and password = $2", name, password)
	return row.Scan()
}

func (s *PostgresStore) Create() {}
