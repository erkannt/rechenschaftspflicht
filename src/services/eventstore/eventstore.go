package eventstore

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	Sequence   int    `json:"sequence"`
	EventType  string `json:"eventType"`
	Tag        string `json:"tag"`
	Comment    string `json:"comment"`
	Value      string `json:"value"`
	RecordedAt string `json:"recordedAt"`
	RecordedBy string `json:"recordedBy"`
}

// EventPayload is used for non-legacy events to store structured data.
type EventPayload struct {
	OriginalEventSequence int `json:"originalEventSequence,omitempty"`
}

type EventStore interface {
	RaiseEvent(event Event) error
	GetAllEvents() ([]Event, error)
	GetAllTags() ([]string, error)
}

type SQLiteEventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) EventStore {
	return &SQLiteEventStore{db: db}
}

// RaiseEvent stores an event in the database.
// For legacy EventRecorded events, it uses the structured columns (tag, comment, value).
// For other event types, it stores the payload as JSON in the value column.
func (s *SQLiteEventStore) RaiseEvent(event Event) error {
	eventType := event.EventType
	if eventType == "" {
		eventType = "EventRecorded"
	}

	// For legacy EventRecorded events, use structured columns
	// For other event types, the payload should already be JSON in event.Value
	stmt := `INSERT INTO events (event_type, tag, comment, value, recordedAt, recordedBy) VALUES (?, ?, ?, ?, ?, ?);`
	_, err := s.db.Exec(stmt, eventType, event.Tag, event.Comment, event.Value, event.RecordedAt, event.RecordedBy)
	return err
}

// GetAllEvents returns all events ordered by sequence (descending).
// It includes the sequence number which can be used to reference events.
func (s *SQLiteEventStore) GetAllEvents() ([]Event, error) {
	rows, err := s.db.Query(`
		SELECT e.sequence, e.tag, e.comment, e.value, e.recordedAt, e.event_type, u.username
		FROM events e
		LEFT JOIN users u ON e.recordedBy = u.email
		ORDER BY e.sequence DESC;
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
		if err := rows.Scan(&e.Sequence, &e.Tag, &e.Comment, &e.Value, &e.RecordedAt, &e.EventType, &e.RecordedBy); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, err
}

func (s *SQLiteEventStore) GetAllTags() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT tag FROM events ORDER BY tag;`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}
