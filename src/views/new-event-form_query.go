package views

import "strings"

// TagSuggestion is the read model for tag suggestions in the new event form.
type TagSuggestion struct {
	Tag string
}

// FormState carries submitted form values and validation errors
// for re-rendering the form after a validation failure.
type FormState struct {
	Tag     string
	Value   string
	Comment string
	Errors  map[string]string // field name → error message
}

// HasErrors returns true if there are validation errors.
func (fs FormState) HasErrors() bool {
	return len(fs.Errors) > 0
}

// ErrorFor returns the error message for a given field, or empty string if none.
func (fs FormState) ErrorFor(field string) string {
	return fs.Errors[field]
}

// GetTagSuggestions retrieves all unique tags from EventRecorded events, sorted alphabetically.
// This is a query-layer concern, so it derives tags from GetAllEvents() rather than
// having a dedicated method in the EventStore interface.
func (h *QueryHandler) GetTagSuggestions() ([]TagSuggestion, error) {
	events, err := h.eventStore.GetAllEvents()
	if err != nil {
		return nil, err
	}

	// Build unique sorted tags from EventRecorded events
	tagSet := make(map[string]bool)
	for _, e := range events {
		if e.EventType == "EventRecorded" && e.Tag != "" {
			tagSet[strings.ToLower(e.Tag)] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	// Sort alphabetically
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i] > tags[j] {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	suggestions := make([]TagSuggestion, 0, len(tags))
	for _, tag := range tags {
		suggestions = append(suggestions, TagSuggestion{Tag: tag})
	}

	return suggestions, nil
}
