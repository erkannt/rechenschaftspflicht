package views

import "strconv"

// PlotData is the read model for the plots view.
// It contains only numeric events that can be visualized.
type PlotData struct {
	Tag        string
	Value      float64
	RecordedAt string
}

// GetEventsForPlots retrieves events and filters/projects them to PlotData read models.
// Only events with valid numeric values are included.
func (h *QueryHandler) GetEventsForPlots() ([]PlotData, error) {
	events, err := h.eventStore.GetAllEvents()
	if err != nil {
		return nil, err
	}

	data := make([]PlotData, 0)
	for _, e := range events {
		// Skip events without a value
		if e.Value == "" {
			continue
		}

		// Skip events with non-numeric values
		valueNum, err := strconv.ParseFloat(e.Value, 64)
		if err != nil {
			continue
		}

		data = append(data, PlotData{
			Tag:        e.Tag,
			Value:      valueNum,
			RecordedAt: e.RecordedAt,
		})
	}

	return data, nil
}
