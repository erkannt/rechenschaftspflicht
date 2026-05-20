package views

// EventListItem is the read model for the all-events list view.
// It contains only the data needed for displaying events in a list.
type EventListItem struct {
	Sequence          int
	RecordedBy        string
	Tag               string
	Value             string
	RecordedAt        string
	Comment           string
	CanMarkAsIncorrect bool
}

// GetAllEventsForList retrieves all events, filters out inactive ones,
// and projects them to EventListItem read models.
func (h *QueryHandler) GetAllEventsForList() ([]EventListItem, error) {
	events, err := h.eventStore.GetAllEvents()
	if err != nil {
		return nil, err
	}

	active := OnlyActiveEvents(events, h.logger)

	items := make([]EventListItem, 0, len(active))
	for _, e := range active {
		items = append(items, EventListItem{
			Sequence:   e.Sequence,
			RecordedBy: e.RecordedBy,
			Tag:        e.Tag,
			Value:      e.Value,
			RecordedAt: e.RecordedAt,
			Comment:    e.Comment,
		})
	}

	return items, nil
}

// GetAllEventsForListWithCurrentUser retrieves all events, filters out inactive ones,
// and projects them to EventListItem read models with CanMarkAsIncorrect set for
// events recorded by the current user.
func (h *QueryHandler) GetAllEventsForListWithCurrentUser(currentUserEmail string) ([]EventListItem, error) {
	items, err := h.GetAllEventsForList()
	if err != nil {
		return nil, err
	}

	for i := range items {
		items[i].CanMarkAsIncorrect = items[i].RecordedBy == currentUserEmail
	}

	return items, nil
}
