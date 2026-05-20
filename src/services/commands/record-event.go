package commands

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// Tag pattern: lowercase letters and hyphens, must start with a letter
var tagPattern = regexp.MustCompile(`^[a-z][a-z-]*$`)

// RecordEvent records a new event with the given data.
// It validates the tag format and value (if provided) before recording.
func (h *CommandHandler) RecordEvent(
	ctx context.Context,
	tag string,
	comment string,
	value string,
	recordedBy string,
) error {
	// Validate tag
	if tag == "" {
		return fmt.Errorf("tag is required")
	}
	if !tagPattern.MatchString(tag) {
		return fmt.Errorf("tag must start with a lowercase letter and contain only lowercase letters and hyphens")
	}

	// Validate value (if provided, must be numeric)
	if value != "" {
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("value must be a valid number: %w", err)
		}
	}

	event := eventstore.Event{
		EventType:  "EventRecorded",
		Tag:        tag,
		Comment:    comment,
		Value:      value,
		RecordedAt: time.Now().Format(time.RFC3339),
		RecordedBy: recordedBy,
	}

	if err := h.eventStore.RaiseEvent(event); err != nil {
		return fmt.Errorf("failed to record event: %w", err)
	}

	h.logger.DebugContext(ctx, "event recorded",
		"tag", tag,
		"recordedBy", recordedBy,
	)

	return nil
}
