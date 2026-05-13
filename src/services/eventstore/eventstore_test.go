package eventstore

import (
	"database/sql"
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/config"
	database "github.com/erkannt/rechenschaftspflicht/services/db"
)

func TestGetAllTags_ReturnsDistinctTagsOrdered(t *testing.T) {
	db := setupTestDB(t)
	store := NewEventStore(db)

	insertEvents(t, db, []Event{
		{Tag: "alpha", Comment: "first"},
		{Tag: "beta", Comment: "second"},
		{Tag: "alpha", Comment: "third"},
		{Tag: "gamma", Comment: "fourth"},
	})

	tags, err := store.GetAllTags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"alpha", "beta", "gamma"}
	assertTagsEqual(t, tags, want)
}

func TestGetAllTags_ReturnsEmptyWhenNoEvents(t *testing.T) {
	db := setupTestDB(t)
	store := NewEventStore(db)

	tags, err := store.GetAllTags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tags) != 0 {
		t.Fatalf("expected empty tags, got %v", tags)
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	cfg := config.Config{SqlitePath: ":memory:"}
	db, err := database.InitDB(cfg)
	if err != nil {
		t.Fatalf("failed to init test DB: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func insertEvents(t *testing.T, db *sql.DB, events []Event) {
	t.Helper()

	stmt := `INSERT INTO events (tag, comment, value, recordedAt, recordedBy) VALUES (?, ?, ?, ?, ?);`
	for _, e := range events {
		_, err := db.Exec(stmt, e.Tag, e.Comment, e.Value, e.RecordedAt, e.RecordedBy)
		if err != nil {
			t.Fatalf("failed to insert event: %v", err)
		}
	}
}

func assertTagsEqual(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d tags, got %d: got=%v, want=%v", len(want), len(got), got, want)
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("tag[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}
