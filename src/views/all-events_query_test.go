package views

import (
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

func TestGetAllEventsForList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		events   []eventstore.Event
		expected []EventListItem
	}{
		{
			name:     "empty event store returns empty list",
			events:   []eventstore.Event{},
			expected: []EventListItem{},
		},
		{
			name: "single event is projected correctly",
			events: []eventstore.Event{
				{
					Tag:        "test-tag",
					Comment:    "test comment",
					Value:      "42",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
			expected: []EventListItem{
				{
					Tag:        "test-tag",
					Comment:    "test comment",
					Value:      "42",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
		},
		{
			name: "multiple events are projected in order",
			events: []eventstore.Event{
				{
					Tag:        "first",
					Comment:    "first comment",
					Value:      "1",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user1@example.com",
				},
				{
					Tag:        "second",
					Comment:    "second comment",
					Value:      "2",
					RecordedAt: "2024-01-02T00:00:00Z",
					RecordedBy: "user2@example.com",
				},
			},
			expected: []EventListItem{
				{
					Tag:        "first",
					Comment:    "first comment",
					Value:      "1",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user1@example.com",
				},
				{
					Tag:        "second",
					Comment:    "second comment",
					Value:      "2",
					RecordedAt: "2024-01-02T00:00:00Z",
					RecordedBy: "user2@example.com",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := &mockEventStore{events: tt.events}
			qh := NewQueryHandler(mockStore)

			result, err := qh.GetAllEventsForList()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d events, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("event %d: expected %+v, got %+v", i, expected, result[i])
				}
			}
		})
	}
}
