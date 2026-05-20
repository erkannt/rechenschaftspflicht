package views

// TagSuggestion is the read model for tag suggestions in the new event form.
type TagSuggestion struct {
	Tag string
}

// GetTagSuggestions retrieves all unique tags from the event store, sorted alphabetically.
func (h *QueryHandler) GetTagSuggestions() ([]TagSuggestion, error) {
	tags, err := h.eventStore.GetAllTags()
	if err != nil {
		return nil, err
	}

	suggestions := make([]TagSuggestion, 0, len(tags))
	for _, tag := range tags {
		suggestions = append(suggestions, TagSuggestion{Tag: tag})
	}

	return suggestions, nil
}
