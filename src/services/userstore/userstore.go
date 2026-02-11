package userstore

import (
	"database/sql"
)

type UserStore interface {
	IsUser(email string) (bool, error)
	AddUser(email string, username string) error
}

type SQLiteUserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) UserStore {
	return &SQLiteUserStore{db: db}
}

func (s *SQLiteUserStore) IsUser(email string) (bool, error) {
	const query = `
		SELECT COUNT(1)
		FROM users
		WHERE LOWER(email) = LOWER(?);
	`

	var count int
	if err := s.db.QueryRow(query, email).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SQLiteUserStore) AddUser(email string, username string) error {
	const query = `
		INSERT INTO users (email, username)
		VALUES (LOWER(?), ?);
	`

	_, err := s.db.Exec(query, email, username)
	return err
}
