package commands

import (
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// mockEventStore is a test double for eventstore.EventStore
type mockEventStore struct {
	recorded []eventstore.Event
}

func (m *mockEventStore) RaiseEvent(event eventstore.Event) error {
	m.recorded = append(m.recorded, event)
	return nil
}

func (m *mockEventStore) GetAllEvents() ([]eventstore.Event, error) {
	return m.recorded, nil
}

func TestNewCommandHandler(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{}
	ch := NewCommandHandler(mockStore, nil)

	if ch == nil {
		t.Fatal("expected CommandHandler to be created, got nil")
	}

	if ch.eventStore != mockStore {
		t.Error("expected eventStore to be set to mockStore")
	}
}
