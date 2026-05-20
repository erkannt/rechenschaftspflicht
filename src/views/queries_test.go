package views

import (
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// mockEventStore is a test double for eventstore.EventStore
type mockEventStore struct {
	events []eventstore.Event
	tags   []string // If set, used by GetAllTags; otherwise derived from events
}

func (m *mockEventStore) RaiseEvent(event eventstore.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventStore) GetAllEvents() ([]eventstore.Event, error) {
	return m.events, nil
}

func (m *mockEventStore) GetAllTags() ([]string, error) {
	// If tags are explicitly set, return them
	if m.tags != nil {
		return m.tags, nil
	}
	// Otherwise, derive unique sorted tags from events
	tagSet := make(map[string]struct{})
	for _, e := range m.events {
		tagSet[e.Tag] = struct{}{}
	}
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	// Sort alphabetically
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i] > tags[j] {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}
	return tags, nil
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
