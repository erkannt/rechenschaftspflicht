package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// ErrNotOwner is returned when a user attempts to mark an event they do not own.
var ErrNotOwner = errors.New("event does not belong to user")

// MarkEventAsIncorrect marks an existing event as incorrect by raising
// an EventMarkedAsIncorrect event.
//
// The originalSequence must refer to an existing EventRecorded event.
// The markedBy parameter should be the email of the user marking the event.
func (h *CommandHandler) MarkEventAsIncorrect(
	ctx context.Context,
	originalSequence int,
	markedBy string,
) error {
	// Validate markedBy
	if markedBy == "" {
		return fmt.Errorf("markedBy is required")
	}

	// Validate that the original event exists
	events, err := h.eventStore.GetAllEvents()
	if err != nil {
		return fmt.Errorf("failed to retrieve events: %w", err)
	}

	var targetEvent *eventstore.Event
	for _, e := range events {
		if e.Sequence == originalSequence {
			targetEvent = &e
			break
		}
	}

	if targetEvent == nil {
		return fmt.Errorf("event with sequence %d does not exist", originalSequence)
	}

	if targetEvent.RecordedBy != markedBy {
		return ErrNotOwner
	}

	// Create the EventMarkedAsIncorrect event
	payload := eventstore.EventPayload{
		OriginalEventSequence: originalSequence,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := eventstore.Event{
		EventType:  "EventMarkedAsIncorrect",
		Value:      string(payloadJSON),
		RecordedAt: time.Now().Format(time.RFC3339),
		RecordedBy: markedBy,
	}

	if err := h.eventStore.RaiseEvent(event); err != nil {
		return fmt.Errorf("failed to raise EventMarkedAsIncorrect: %w", err)
	}

	h.logger.DebugContext(ctx, "event marked as incorrect",
		"originalSequence", originalSequence,
		"markedBy", markedBy,
	)

	return nil
}
