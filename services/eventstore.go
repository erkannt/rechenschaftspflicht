package services

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	Tag       string `json:"tag"`
	Comment   string `json:"comment"`
	Value     string `json:"value"`
	CreatedAt string `json:"createdAt"`
}

type EventStore interface {
	Record(event Event) error
	GetAll() ([]Event, error)
}

type SQLiteEventStore struct {
	db *sql.DB
}

func NewEventStore() EventStore {
	// Ensure the data directory exists
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		panic(err)
	}

	dbPath := filepath.Join("data", "state.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	// Create the events table if it does not exist
	createTable := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tag TEXT,
		comment TEXT,
		value TEXT,
		createdAt TEXT
	);
	`
	if _, err = db.Exec(createTable); err != nil {
		panic(err)
	}

	return &SQLiteEventStore{db: db}
}

func (s *SQLiteEventStore) Record(event Event) error {
	stmt := `INSERT INTO events (tag, comment, value, createdAt) VALUES (?, ?, ?, ?);`
	_, err := s.db.Exec(stmt, event.Tag, event.Comment, event.Value, event.CreatedAt)
	return err
}

func (s *SQLiteEventStore) GetAll() ([]Event, error) {
	rows, err := s.db.Query(`SELECT tag, comment, value, createdAt FROM events ORDER BY id;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.Tag, &e.Comment, &e.Value, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
