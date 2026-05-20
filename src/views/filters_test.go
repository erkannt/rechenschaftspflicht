package views

import (
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

func TestOnlyActiveEvents_ExcludesMarkedIncorrectEvents(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()
	events := []eventstore.Event{
		{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", Value: "20.5"},
		{Sequence: 2, EventType: "EventRecorded", Tag: "humidity", Value: "65"},
		{Sequence: 3, EventType: "EventMarkedAsIncorrect", Value: `{"originalEventSequence":1}`},
	}

	result := OnlyActiveEvents(events, logger)

	if len(result) != 1 {
		t.Fatalf("expected 1 active event, got %d", len(result))
	}
	if result[0].Sequence != 2 {
		t.Errorf("expected sequence 2, got %d", result[0].Sequence)
	}
}

func TestOnlyActiveEvents_KeepsUnmarkedEvents(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()
	events := []eventstore.Event{
		{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", Value: "20.5"},
		{Sequence: 2, EventType: "EventRecorded", Tag: "humidity", Value: "65"},
		{Sequence: 3, EventType: "EventRecorded", Tag: "pressure", Value: "1013"},
	}

	result := OnlyActiveEvents(events, logger)

	if len(result) != 3 {
		t.Fatalf("expected 3 active events, got %d", len(result))
	}
}

func TestOnlyActiveEvents_ExcludesMetadataEvents(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()
	events := []eventstore.Event{
		{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", Value: "20.5"},
		{Sequence: 2, EventType: "EventMarkedAsIncorrect", Value: `{"originalEventSequence":1}`},
	}

	result := OnlyActiveEvents(events, logger)

	if len(result) != 0 {
		t.Fatalf("expected 0 active events (all marked), got %d", len(result))
	}
}

func TestOnlyActiveEvents_EmptyInput(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()
	events := []eventstore.Event{}

	result := OnlyActiveEvents(events, logger)

	if len(result) != 0 {
		t.Fatalf("expected 0 events for empty input, got %d", len(result))
	}
}

func TestOnlyActiveEvents_MultipleMarksOnSameEvent(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()
	events := []eventstore.Event{
		{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", Value: "20.5"},
		{Sequence: 2, EventType: "EventMarkedAsIncorrect", Value: `{"originalEventSequence":1}`},
		{Sequence: 3, EventType: "EventMarkedAsIncorrect", Value: `{"originalEventSequence":1}`},
	}

	result := OnlyActiveEvents(events, logger)

	if len(result) != 0 {
		t.Fatalf("expected 0 active events, got %d", len(result))
	}
}

func TestOnlyActiveEvents_MalformedJsonPayload(t *testing.T) {
	t.Parallel()

	logger := newTestLogger()
	events := []eventstore.Event{
		{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", Value: "20.5"},
		{Sequence: 2, EventType: "EventMarkedAsIncorrect", Value: `invalid json`},
	}

	result := OnlyActiveEvents(events, logger)

	// Malformed JSON should be ignored, so event 1 should still be active
	if len(result) != 1 {
		t.Fatalf("expected 1 active event (malformed JSON ignored), got %d", len(result))
	}
	if result[0].Sequence != 1 {
		t.Errorf("expected sequence 1, got %d", result[0].Sequence)
	}
}
