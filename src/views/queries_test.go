package views

import (
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// mockEventStore is a test double for eventstore.EventStore
type mockEventStore struct {
	events []eventstore.Event
}

func (m *mockEventStore) RaiseEvent(event eventstore.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventStore) GetAllEvents() ([]eventstore.Event, error) {
	return m.events, nil
}

func TestNewQueryHandler(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{}
	qh := NewQueryHandler(mockStore, newTestLogger())

	if qh == nil {
		t.Fatal("expected QueryHandler to be created, got nil")
	}

	if qh.eventStore != mockStore {
		t.Error("expected eventStore to be set to mockStore")
	}
}
