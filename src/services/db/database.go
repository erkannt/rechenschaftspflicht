package database

import (
	"database/sql"
	"os"

	"github.com/erkannt/rechenschaftspflicht/services/config"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(config config.Config) (*sql.DB, error) {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return nil, err
	}

	createEventsTable := `
	CREATE TABLE IF NOT EXISTS events (
		sequence INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL DEFAULT 'EventRecorded',
		tag TEXT,
		comment TEXT,
		value TEXT,
		recordedAt TEXT,
		recordedBy TEXT
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

	// Migration: Add event_type column if it doesn't exist (for existing databases)
	// SQLite doesn't support ADD COLUMN IF NOT EXISTS, so we check manually
	migrateAddEventType := `
	SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='event_type';
	`
	var count int
	err = db.QueryRow(migrateAddEventType).Scan(&count)
	if err == nil && count == 0 {
		_, err = db.Exec(`ALTER TABLE events ADD COLUMN event_type TEXT NOT NULL DEFAULT 'EventRecorded';`)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
