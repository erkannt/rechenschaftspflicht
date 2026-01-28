package eventstore

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	Tag        string `json:"tag"`
	Comment    string `json:"comment"`
	Value      string `json:"value"`
	RecordedAt string `json:"recordedAt"`
	RecordedBy string `json:"recordedBy"`
}

type EventStore interface {
	Record(event Event) error
	GetAll() ([]Event, error)
}

type SQLiteEventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) EventStore {
	return &SQLiteEventStore{db: db}
}

func (s *SQLiteEventStore) Record(event Event) error {
	stmt := `INSERT INTO events (tag, comment, value, recordedAt, recordedBy) VALUES (?, ?, ?, ?, ?);`
	_, err := s.db.Exec(stmt, event.Tag, event.Comment, event.Value, event.RecordedAt, event.RecordedBy)
	return err
}

func (s *SQLiteEventStore) GetAll() ([]Event, error) {
	rows, err := s.db.Query(`
		SELECT e.tag, e.comment, e.value, e.recordedAt, u.username
		FROM events e
		LEFT JOIN users u ON e.recordedBy = u.email
		ORDER BY e.recordedAt DESC;
	`)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.Tag, &e.Comment, &e.Value, &e.RecordedAt, &e.RecordedBy); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, err
}
