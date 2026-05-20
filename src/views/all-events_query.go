package views

// EventListItem is the read model for the all-events list view.
// It contains only the data needed for displaying events in a list.
type EventListItem struct {
	RecordedBy string
	Tag        string
	Value      string
	RecordedAt string
	Comment    string
}

// GetAllEventsForList retrieves all events and projects them to EventListItem read models.
func (h *QueryHandler) GetAllEventsForList() ([]EventListItem, error) {
	events, err := h.eventStore.GetAllEvents()
	if err != nil {
		return nil, err
	}

	items := make([]EventListItem, 0, len(events))
	for _, e := range events {
		items = append(items, EventListItem{
			RecordedBy: e.RecordedBy,
			Tag:        e.Tag,
			Value:      e.Value,
			RecordedAt: e.RecordedAt,
			Comment:    e.Comment,
		})
	}

	return items, nil
}
