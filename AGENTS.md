# AGENTS.md

This file contains guidelines and commands for agentic coding agents working on this Go web application.

## Project Overview

This is a Go web application called "Rechenschaftspflicht" that uses:

- **Go 1.25.6** with standard library HTTP routing
- **templ** for HTML templating (generated files in `src/views/*_templ.go`)
- **SQLite** for data persistence
- **JWT** for authentication
- **httplrouter** for HTTP routing
- **Docker** for containerization

## Build/Development Commands

IMPORTANT: prefix all shell commands with `PATH=$PATH:~/mise/shims` so that dependencies are available.

All make commands must be run from the repository root (same directory as the Makefile), not from inside `src/`.

### Validating changes

```bash
make check                 # Run go build, go vet, and golangci-lint
make fix                   # Auto-format and fix code (go fmt, go fix)
make test                  # Run unit tests
make generate              # generate go code from templ files to manually check if templ file works as expected
```

## Code Style Guidelines

### Import Organization

- Group imports in three blocks: standard library, third-party, local packages
- Use blank import `_ "github.com/mattn/go-sqlite3"` for database drivers
- Local imports use full module path: `github.com/erkannt/rechenschaftspflicht/...`

```go
import (
    "context"
    "fmt"
    "log/slog"

    "github.com/julienschmidt/httprouter"

    "github.com/erkannt/rechenschaftspflicht/services/authentication"
    database "github.com/erkannt/rechenschaftspflicht/services/db"
)
```

### Naming Conventions

- **Packages**: lowercase, single words when possible (`handlers`, `services`)
- **Files**: snake_case for files (but prefer Go package structure)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase, abbreviate common patterns (`ctx`, `w`, `r`, `err`)
- **Constants**: PascalCase or ALL_CAPS for exported
- **Interfaces**: End with `er` suffix (`EventStore`, `Auth`)
- **Structs**: PascalCase, exported fields

### Error Handling

- Always handle errors immediately after function calls
- Use structured error wrapping with `%w` verb
- Return errors from functions unless handling them completely
- Use `fmt.Errorf("context: %w", err)` for error wrapping
- Log errors with context using structured logging (`slog`)

```go
db, err := database.InitDB()
if err != nil {
    return fmt.Errorf("could not init database: %w", err)
}
```

### HTTP Handler Pattern

- Use httprouter handlers: `func(w http.ResponseWriter, r *http.Request, _ httprouter.Params)`
- Return `httprouter.Handle` from higher-order handlers for dependency injection
- Always check for render errors and return HTTP 500 with logging

```go
func RecordEventPostHandler(eventStore eventstore.EventStore, auth authentication.Auth) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
        // Handler implementation
    }
}
```

### Database Pattern

- Use interface-based design for stores
- Parameterize all queries to prevent SQL injection
- Always close rows with defer
- Handle errors appropriately and return structured errors

```go
func (s *SQLiteEventStore) GetAll() ([]Event, error) {
    rows, err := s.db.Query(`SELECT ...`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    // Process rows...
}
```

### Templ Components

- Edit `.templ` files in `src/views/` directory
- Never edit generated `*_templ.go` files directly
- Use `@Component()` syntax for component composition
- Export templ components that are used across packages

### Service Layer Architecture

- Services are in `src/services/` with subdirectories per domain
- Each service has a clear interface and concrete implementation
- Dependency injection happens in `main.go` and passed through handlers
- Services handle business logic, handlers handle HTTP concerns

### Logging

- Use structured logging with `log/slog`
- Log at appropriate levels (Debug, Info, Warn, Error)
- Include context in log messages (user ID, request ID, etc.)
- Avoid `fmt.Printf` for logging - use proper logger

### Context Usage

- Pass `context.Context` through the call chain
- Use context for cancellation and timeouts
- Store context in request handlers: `r.Context()`

### Code Organization

- **main.go**: Application entry point and dependency setup
- **handlers/**: HTTP request handlers
- **services/**: Business logic and data access
- **views/**: Templ templates and generated HTML
- **routes.go**: HTTP routing configuration

### Static Files

- Static assets in `src/assets/` directory
- Embedded in binary using `//go:embed`
- Served at `/assets/*` path

### Environment Configuration

- `Config` struct in config service documents the configuration
- Use `.env` file for development (copy from `.env.example`)
- Configuration loaded via custom reflection-based parser in `src/services/config/env/`
- Environment variables need to be kept up to date in `docker-compose.yaml`

### Security

- Use JWT tokens for authentication
- Implement CSRF protection where appropriate
- Validate all input (form data, JSON, URL params)
- Use parameterized queries to prevent SQL injection
- Set appropriate HTTP headers (Content-Type, etc.)

### Performance

- Use connection pooling for database
- Implement proper graceful shutdown
- Consider caching for frequently accessed data
- Use context timeouts for external calls

## Event Sourcing

The application uses an event-sourced architecture with a single `events` table.

### Event Types

- **EventRecorded** (legacy): Structured data in separate columns (tag, comment, value)
- **EventMarkedAsIncorrect**: JSON payload in `value` column with `originalEventSequence`

### EventStore Interface

The EventStore exposes two minimal methods:

```go
type EventStore interface {
    RaiseEvent(event Event) error      // Store any event type
    GetAllEvents() ([]Event, error)    // Retrieve all events with sequence numbers
}
```

Query-specific operations (like tag suggestions) are implemented in the query layer
by composing `GetAllEvents()` with filtering/projection logic.

### Filtering Events

Use `OnlyActiveEvents()` in the views package to filter out events marked as incorrect:

```go
events, err := h.eventStore.GetAllEvents()
active := views.OnlyActiveEvents(events, h.logger)
// Project active events to view model
```

### Adding New Event Types

1. Define the event type string (e.g., `"EventMarkedAsIncorrect"`)
2. Store payload as JSON in the `value` column
3. Update `OnlyActiveEvents()` filter if the event should affect visibility
4. Create a command in `services/commands/` to raise the event

## Feature: Mark Events as Incorrect

Users can mark their recorded events as incorrect via the All Events page.

### How It Works

1. User clicks "Mark Incorrect" button next to an event
2. POST request to `/mark-event-incorrect` with event sequence
3. `EventMarkedAsIncorrect` event is raised with JSON payload
4. All queries filter out the marked event via `OnlyActiveEvents()`

### Files Involved

- `src/services/commands/mark-event-incorrect.go` - Command implementation
- `src/handlers/events.go` - HTTP handler (`MarkEventIncorrectPostHandler`)
- `src/views/filters.go` - `OnlyActiveEvents()` filter function
- `src/views/all-events.templ` - UI with Mark Incorrect button
