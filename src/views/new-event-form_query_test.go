package views

import (
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

func TestGetTagSuggestions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		events   []eventstore.Event
		expected []TagSuggestion
	}{
		{
			name:     "empty event store returns empty list",
			events:   []eventstore.Event{},
			expected: []TagSuggestion{},
		},
		{
			name: "single tag is returned",
			events: []eventstore.Event{
				{
					Tag:        "temperature",
					Comment:    "first reading",
					Value:      "20",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
			expected: []TagSuggestion{
				{Tag: "temperature"},
			},
		},
		{
			name: "duplicate tags are deduplicated",
			events: []eventstore.Event{
				{
					Tag:        "temperature",
					Comment:    "first",
					Value:      "20",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
				{
					Tag:        "temperature",
					Comment:    "second",
					Value:      "21",
					RecordedAt: "2024-01-02T00:00:00Z",
					RecordedBy: "user@example.com",
				},
				{
					Tag:        "humidity",
					Comment:    "first",
					Value:      "60",
					RecordedAt: "2024-01-03T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
			expected: []TagSuggestion{
				{Tag: "humidity"},
				{Tag: "temperature"},
			},
		},
		{
			name: "tags are sorted alphabetically",
			events: []eventstore.Event{
				{
					Tag:        "zebra",
					Comment:    "z",
					Value:      "1",
					RecordedAt: "2024-01-01T00:00:00Z",
					RecordedBy: "user@example.com",
				},
				{
					Tag:        "alpha",
					Comment:    "a",
					Value:      "2",
					RecordedAt: "2024-01-02T00:00:00Z",
					RecordedBy: "user@example.com",
				},
				{
					Tag:        "beta",
					Comment:    "b",
					Value:      "3",
					RecordedAt: "2024-01-03T00:00:00Z",
					RecordedBy: "user@example.com",
				},
			},
			expected: []TagSuggestion{
				{Tag: "alpha"},
				{Tag: "beta"},
				{Tag: "zebra"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := &mockEventStore{events: tt.events}
			qh := NewQueryHandler(mockStore)

			result, err := qh.GetTagSuggestions()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d tags, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("tag %d: expected %+v, got %+v", i, expected, result[i])
				}
			}
		})
	}
}
