package commands

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRecordEvent_ValidEvent(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.RecordEvent(context.Background(), "temperature", "morning reading", "23.5", "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockStore.recorded) != 1 {
		t.Fatalf("expected 1 event to be recorded, got %d", len(mockStore.recorded))
	}

	event := mockStore.recorded[0]
	if event.Tag != "temperature" {
		t.Errorf("expected tag 'temperature', got %q", event.Tag)
	}
	if event.Comment != "morning reading" {
		t.Errorf("expected comment 'morning reading', got %q", event.Comment)
	}
	if event.Value != "23.5" {
		t.Errorf("expected value '23.5', got %q", event.Value)
	}
	if event.RecordedBy != "user@example.com" {
		t.Errorf("expected recordedBy 'user@example.com', got %q", event.RecordedBy)
	}
	if event.EventType != "EventRecorded" {
		t.Errorf("expected eventType 'EventRecorded', got %q", event.EventType)
	}
}

func TestRecordEvent_EmptyTag(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.RecordEvent(context.Background(), "", "comment", "10", "user@example.com")
	if err == nil {
		t.Fatal("expected error for empty tag, got nil")
	}
}

func TestRecordEvent_InvalidTagFormat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		tag  string
	}{
		{"tag with uppercase", "Temperature"},
		{"tag with numbers", "temp1"},
		{"tag starting with number", "1temp"},
		{"tag with underscore", "temp_reading"},
		{"tag with spaces", "temp reading"},
		{"tag starting with hyphen", "-temp"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockStore := &mockEventStore{}
			ch := NewCommandHandler(mockStore, newTestLogger())

			err := ch.RecordEvent(context.Background(), tc.tag, "comment", "10", "user@example.com")
			if err == nil {
				t.Errorf("expected error for tag %q, got nil", tc.tag)
			}
		})
	}
}

func TestRecordEvent_ValidTagFormats(t *testing.T) {
	t.Parallel()

	validTags := []string{"temperature", "temp-reading", "a", "a-b-c", "humidity"}

	for _, tag := range validTags {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()

			mockStore := &mockEventStore{}
			ch := NewCommandHandler(mockStore, newTestLogger())

			err := ch.RecordEvent(context.Background(), tag, "comment", "10", "user@example.com")
			if err != nil {
				t.Errorf("expected no error for tag %q, got %v", tag, err)
			}
		})
	}
}

func TestRecordEvent_InvalidValueFormat(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.RecordEvent(context.Background(), "temperature", "comment", "not-a-number", "user@example.com")
	if err == nil {
		t.Fatal("expected error for non-numeric value, got nil")
	}
}

func TestRecordEvent_EmptyValueIsAllowed(t *testing.T) {
	t.Parallel()

	mockStore := &mockEventStore{}
	ch := NewCommandHandler(mockStore, newTestLogger())

	err := ch.RecordEvent(context.Background(), "note", "just a comment without value", "", "user@example.com")
	if err != nil {
		t.Fatalf("expected no error for empty value, got %v", err)
	}

	if len(mockStore.recorded) != 1 {
		t.Fatalf("expected 1 event to be recorded, got %d", len(mockStore.recorded))
	}

	if mockStore.recorded[0].Value != "" {
		t.Errorf("expected empty value, got %q", mockStore.recorded[0].Value)
	}
}
