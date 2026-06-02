package commands

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

func TestMarkEventAsIncorrect_ValidEvent(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{
		recorded: []eventstore.Event{
			{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", RecordedBy: "user@example.com"},
		},
	}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.MarkEventAsIncorrect(context.Background(), 1, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockStore.recorded) != 2 {
		t.Fatalf("expected 2 events (original + mark), got %d", len(mockStore.recorded))
	}

	markEvent := mockStore.recorded[1]
	if markEvent.EventType != "EventMarkedAsIncorrect" {
		t.Errorf("expected EventType 'EventMarkedAsIncorrect', got %q", markEvent.EventType)
	}
	if markEvent.RecordedBy != "user@example.com" {
		t.Errorf("expected recordedBy 'user@example.com', got %q", markEvent.RecordedBy)
	}

	// Verify payload
	var payload eventstore.EventPayload
	if err := json.Unmarshal([]byte(markEvent.Value), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload.OriginalEventSequence != 1 {
		t.Errorf("expected originalEventSequence 1, got %d", payload.OriginalEventSequence)
	}
}

func TestMarkEventAsIncorrect_NonExistentEvent(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{
		recorded: []eventstore.Event{
			{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", RecordedBy: "user@example.com"},
		},
	}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.MarkEventAsIncorrect(context.Background(), 999, "user@example.com")
	if err == nil {
		t.Fatal("expected error for non-existent event, got nil")
	}
}

func TestMarkEventAsIncorrect_EmptyMarkedBy(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{
		recorded: []eventstore.Event{
			{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", RecordedBy: "user@example.com"},
		},
	}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.MarkEventAsIncorrect(context.Background(), 1, "")
	if err == nil {
		t.Fatal("expected error for empty markedBy, got nil")
	}
}

func TestMarkEventAsIncorrect_Idempotent(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{
		recorded: []eventstore.Event{
			{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", RecordedBy: "user@example.com"},
		},
	}
	ch := NewCommandHandler(mockStore, newTestLogger())

	// Same owner marks their own event twice — should be allowed (creates two mark events)
	err := ch.MarkEventAsIncorrect(context.Background(), 1, "user@example.com")
	if err != nil {
		t.Fatalf("first mark failed: %v", err)
	}

	err = ch.MarkEventAsIncorrect(context.Background(), 1, "user@example.com")
	if err != nil {
		t.Fatalf("second mark failed: %v", err)
	}

	// Should have 3 events now: original + 2 marks
	if len(mockStore.recorded) != 3 {
		t.Fatalf("expected 3 events, got %d", len(mockStore.recorded))
	}
}

func TestMarkEventAsIncorrect_RecordsTimestamp(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{
		recorded: []eventstore.Event{
			{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", RecordedBy: "user@example.com"},
		},
	}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.MarkEventAsIncorrect(context.Background(), 1, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	markEvent := mockStore.recorded[1]
	if markEvent.RecordedAt == "" {
		t.Error("expected RecordedAt to be set")
	}
}

func TestMarkEventAsIncorrect_NotOwnEvent(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{
		recorded: []eventstore.Event{
			{Sequence: 1, EventType: "EventRecorded", Tag: "temperature", RecordedBy: "owner@example.com"},
		},
	}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.MarkEventAsIncorrect(context.Background(), 1, "attacker@example.com")
	if err == nil {
		t.Fatal("expected error when marking another user's event, got nil")
	}
	if !errors.Is(err, ErrNotOwner) {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}

	// No new events should have been recorded
	if len(mockStore.recorded) != 1 {
		t.Errorf("expected 1 event (original only), got %d", len(mockStore.recorded))
	}
}
