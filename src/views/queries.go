package views

import (
	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// QueryHandler provides read projections for views.
// It transforms raw events from the eventstore into view-specific read models.
type QueryHandler struct {
	eventStore eventstore.EventStore
}

// NewQueryHandler creates a new QueryHandler with the given eventstore.
func NewQueryHandler(eventStore eventstore.EventStore) *QueryHandler {
	return &QueryHandler{eventStore: eventStore}
}
