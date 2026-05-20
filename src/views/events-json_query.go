package views

import "strconv"

// EventJson is the read model for the JSON API endpoint.
// It includes both the raw value string and the parsed numeric value.
type EventJson struct {
	Tag        string  `json:"tag"`
	Comment    string  `json:"comment"`
	Value      string  `json:"value"`
	ValueNum   float64 `json:"valueNum"`
	RecordedAt string  `json:"recordedAt"`
	RecordedBy string  `json:"recordedBy"`
}

// GetEventsForJson retrieves events, filters out inactive ones,
// and filters/projects them for JSON API response.
// Only events with valid numeric values are included.
func (h *QueryHandler) GetEventsForJson() ([]EventJson, error) {
	events, err := h.eventStore.GetAllEvents()
	if err != nil {
		return nil, err
	}

	active := OnlyActiveEvents(events, h.logger)

	result := make([]EventJson, 0)
	for _, e := range active {
		// Skip events without a value
		if e.Value == "" {
			continue
		}

		// Skip events with non-numeric values
		valueNum, err := strconv.ParseFloat(e.Value, 64)
		if err != nil {
			continue
		}

		result = append(result, EventJson{
			Tag:        e.Tag,
			Comment:    e.Comment,
			Value:      e.Value,
			ValueNum:   valueNum,
			RecordedAt: e.RecordedAt,
			RecordedBy: e.RecordedBy,
		})
	}

	return result, nil
}
