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
					Sequence:       1,
					EventType:      "EventRecorded",
					Tag:            "test-tag",
					Comment:        "test comment",
					Value:          "42",
					RecordedAt:     "2024-01-01T00:00:00Z",
					RecordedBy:     "user@example.com",
					RecordedByName: "testuser",
				},
			},
			expected: []EventListItem{
				{
					Sequence:        1,
					RecordedBy:      "testuser",
					RecordedByEmail: "user@example.com",
					Tag:             "test-tag",
					Comment:         "test comment",
					Value:           "42",
					RecordedAt:      "2024-01-01T00:00:00Z",
				},
			},
		},
		{
			name: "uppercase tag is lowercased in projection",
			events: []eventstore.Event{
				{
					Sequence:       1,
					EventType:      "EventRecorded",
					Tag:            "My-Tag",
					Comment:        "c",
					Value:          "1",
					RecordedAt:     "2024-01-01T00:00:00Z",
					RecordedBy:     "user@example.com",
					RecordedByName: "u",
				},
			},
			expected: []EventListItem{
				{
					Sequence:        1,
					RecordedBy:      "u",
					RecordedByEmail: "user@example.com",
					Tag:             "my-tag",
					Comment:         "c",
					Value:           "1",
					RecordedAt:      "2024-01-01T00:00:00Z",
				},
			},
		},
		{
			name: "multiple events are projected in order",
			events: []eventstore.Event{
				{
					Sequence:       1,
					EventType:      "EventRecorded",
					Tag:            "first",
					Comment:        "first comment",
					Value:          "1",
					RecordedAt:     "2024-01-01T00:00:00Z",
					RecordedBy:     "user1@example.com",
					RecordedByName: "user1",
				},
				{
					Sequence:       2,
					EventType:      "EventRecorded",
					Tag:            "second",
					Comment:        "second comment",
					Value:          "2",
					RecordedAt:     "2024-01-02T00:00:00Z",
					RecordedBy:     "user2@example.com",
					RecordedByName: "user2",
				},
			},
			expected: []EventListItem{
				{
					Sequence:        1,
					RecordedBy:      "user1",
					RecordedByEmail: "user1@example.com",
					Tag:             "first",
					Comment:         "first comment",
					Value:           "1",
					RecordedAt:      "2024-01-01T00:00:00Z",
				},
				{
					Sequence:        2,
					RecordedBy:      "user2",
					RecordedByEmail: "user2@example.com",
					Tag:             "second",
					Comment:         "second comment",
					Value:           "2",
					RecordedAt:      "2024-01-02T00:00:00Z",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := &mockEventStore{events: tt.events}
			qh := NewQueryHandler(mockStore, newTestLogger())

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

func TestGetAllEventsForList_CurrentUserCanMarkOwnEvents(t *testing.T) {
	t.Parallel()

	currentUser := "current@example.com"
	events := []eventstore.Event{
		{
			Sequence:       1,
			EventType:      "EventRecorded",
			Tag:            "my-event",
			Comment:        "my event",
			Value:          "10",
			RecordedAt:     "2024-01-01T00:00:00Z",
			RecordedBy:     currentUser,
			RecordedByName: "currentuser",
		},
		{
			Sequence:       2,
			EventType:      "EventRecorded",
			Tag:            "other-event",
			Comment:        "other event",
			Value:          "20",
			RecordedAt:     "2024-01-02T00:00:00Z",
			RecordedBy:     "other@example.com",
			RecordedByName: "otheruser",
		},
	}

	mockStore := &mockEventStore{events: events}
	qh := NewQueryHandler(mockStore, newTestLogger())

	result, err := qh.GetAllEventsForListWithCurrentUser(currentUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}

	// First event should be markable by current user
	if !result[0].CanMarkAsIncorrect {
		t.Error("expected current user's event to be markable")
	}

	// Second event should NOT be markable by current user
	if result[1].CanMarkAsIncorrect {
		t.Error("expected other user's event to NOT be markable")
	}
}
