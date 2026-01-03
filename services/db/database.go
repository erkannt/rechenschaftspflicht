package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		return nil, err
	}

	dbPath := filepath.Join("data", "state.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	createEventsTable := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tag TEXT,
		comment TEXT,
		value TEXT,
		createdAt TEXT
	);
	`

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		email TEXT
	);
	`

	if _, err = db.Exec(createEventsTable); err != nil {
		return nil, err
	}
	if _, err = db.Exec(createUsersTable); err != nil {
		return nil, err
	}

	return db, nil
}
