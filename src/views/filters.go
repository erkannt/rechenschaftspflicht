package views

import (
	"encoding/json"
	"log/slog"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// OnlyActiveEvents filters out EventRecorded events that have been
// marked as incorrect, and excludes the EventMarkedAsIncorrect events themselves.
//
// It parses EventMarkedAsIncorrect events to find which original event sequences
// have been marked, then filters out those EventRecorded events.
func OnlyActiveEvents(events []eventstore.Event, logger *slog.Logger) []eventstore.Event {
	// Build set of sequences that have been marked incorrect
	incorrectSequences := make(map[int]bool)
	var filtered []eventstore.Event

	// First pass: identify which sequences are marked incorrect
	for _, e := range events {
		if e.EventType == "EventMarkedAsIncorrect" {
			var payload eventstore.EventPayload
			if err := json.Unmarshal([]byte(e.Value), &payload); err != nil {
				if logger != nil {
					logger.Warn("failed to parse EventMarkedAsIncorrect payload",
						"sequence", e.Sequence,
						"value", e.Value,
						"error", err,
					)
				}
				continue
			}
			incorrectSequences[payload.OriginalEventSequence] = true
		}
	}

	// Second pass: keep only EventRecorded events that are not marked incorrect
	for _, e := range events {
		if e.EventType != "EventRecorded" {
			// Skip metadata events (like EventMarkedAsIncorrect)
			continue
		}
		if incorrectSequences[e.Sequence] {
			// Skip events that have been marked incorrect
			continue
		}
		filtered = append(filtered, e)
	}

	return filtered
}
