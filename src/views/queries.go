package views

import (
	"log/slog"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// QueryHandler provides read projections for views.
// It transforms raw events from the eventstore into view-specific read models.
type QueryHandler struct {
	eventStore eventstore.EventStore
	logger     *slog.Logger
}

// NewQueryHandler creates a new QueryHandler with the given eventstore.
func NewQueryHandler(eventStore eventstore.EventStore, logger *slog.Logger) *QueryHandler {
	return &QueryHandler{
		eventStore: eventStore,
		logger:     logger,
	}
}
