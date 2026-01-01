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
	if _, err = db.Exec(createEventsTable); err != nil {
		return nil, err
	}

	return db, nil
}

func NewEventStore(db *sql.DB) EventStore {
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
