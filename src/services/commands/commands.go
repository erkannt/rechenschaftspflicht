package commands

import (
	"log/slog"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
)

// CommandHandler handles write operations (commands) for the application.
type CommandHandler struct {
	eventStore eventstore.EventStore
	logger     *slog.Logger
}

// NewCommandHandler creates a new CommandHandler with the given dependencies.
func NewCommandHandler(eventStore eventstore.EventStore, logger *slog.Logger) *CommandHandler {
	return &CommandHandler{
		eventStore: eventStore,
		logger:     logger,
	}
}
