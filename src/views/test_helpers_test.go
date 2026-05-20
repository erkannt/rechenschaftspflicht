package views

import (
	"log/slog"
	"os"
)

// newTestLogger creates a logger for tests that only outputs errors.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}
