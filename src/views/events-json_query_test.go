package views

import (
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

func TestGetEventsForJson(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		events   []eventstore.Event
		expected []EventJson
	}{
		{
			name:     "empty event store returns empty list",
			events:   []eventstore.Event{},
			expected: []EventJson{},
		},
		{
			name: "event without value is skipped",
			events: []eventstore.Event{
				{
					Sequence:   1,
					EventType:  "EventRecorded",
					Tag:        "no-value",
					Comment:    "no value here",
					Value:      "",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
			expected: []EventJson{},
		},
		{
			name: "event with valid numeric value is included",
			events: []eventstore.Event{
				{
					Sequence:   1,
					EventType:  "EventRecorded",
					Tag:        "temperature",
					Comment:    "morning reading",
					Value:      "23.5",
					RecordedAt: "2024-01-01T08:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
			expected: []EventJson{
				{
					Sequence:   1,
					Tag:        "temperature",
					Comment:    "morning reading",
					Value:      "23.5",
					ValueNum:   23.5,
					RecordedAt: "2024-01-01T08:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
		},
		{
			name: "event with invalid numeric value is skipped",
			events: []eventstore.Event{
				{
					Sequence:   1,
					EventType:  "EventRecorded",
					Tag:        "invalid",
					Comment:    "not a number",
					Value:      "not-a-number",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
			expected: []EventJson{},
		},
		{
			name: "uppercase tag is lowercased",
			events: []eventstore.Event{
				{
					Sequence:   1,
					EventType:  "EventRecorded",
					Tag:        "Temperature",
					Value:      "20",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
			expected: []EventJson{
				{
					Sequence:   1,
					Tag:        "temperature",
					Value:      "20",
					ValueNum:   20.0,
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
		},
		{
			name: "multiple events filtered and projected correctly",
			events: []eventstore.Event{
				{
					Sequence:   1,
					EventType:  "EventRecorded",
					Tag:        "temperature",
					Comment:    "valid",
					Value:      "20.0",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user1@example.com",
				},
				{
					Sequence:   2,
					EventType:  "EventRecorded",
					Tag:        "humidity",
					Comment:    "empty value",
					Value:      "",
					RecordedAt: "2024-01-02T00:00:00Z",
					RecordedBy: "user2@example.com",
				},
				{
					Sequence:   3,
					EventType:  "EventRecorded",
					Tag:        "pressure",
					Comment:    "valid",
					Value:      "1013.25",
					RecordedAt: "2024-01-03T00:00:00Z",
					RecordedBy: "user3@example.com",
				},
			},
			expected: []EventJson{
				{
					Sequence:   1,
					Tag:        "temperature",
					Comment:    "valid",
					Value:      "20.0",
					ValueNum:   20.0,
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user1@example.com",
				},
				{
					Sequence:   3,
					Tag:        "pressure",
					Comment:    "valid",
					Value:      "1013.25",
					ValueNum:   1013.25,
					RecordedAt: "2024-01-03T00:00:00Z",
					RecordedBy: "user3@example.com",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := &mockEventStore{events: tt.events}
			qh := NewQueryHandler(mockStore, newTestLogger())

			result, err := qh.GetEventsForJson()
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
