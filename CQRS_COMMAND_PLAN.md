# CQRS Command Layer Implementation Plan

## Goals

1. **Proper CQRS Separation**: Commands in `src/services/commands/`, Queries in `src/views/`
2. **Simple Return Values**: No result structs, return errors or nil
3. **Event Versioning Support**: Prepare for corrections/updates (EventRecorded, EventMarkedAsIncorrect, EventUpdated)
4. **Expand-and-Contract**: Safe schema migration with zero downtime

---

## Database Schema Evolution

### Current Schema
```sql
events (
    sequence INTEGER PRIMARY KEY AUTOINCREMENT,
    tag TEXT,
    comment TEXT,
    value TEXT,
    recordedAt TEXT,
    recordedBy TEXT
)
```

### Target Schema (Expand Phase)
```sql
events (
    sequence INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL DEFAULT 'EventRecorded',  -- NEW
    tag TEXT,
    comment TEXT,
    value TEXT,
    recordedAt TEXT,
    recordedBy TEXT,
    corrects_event_id INTEGER REFERENCES events(sequence)  -- NEW, nullable
)
```

### Migration Strategy

| Phase | Action | Risk |
|-------|--------|------|
| **Expand** | Add columns with defaults | Low - existing rows get defaults automatically |
| **Migrate** | Update code to use new columns | Medium - test thoroughly |
| **Contract** | Add indexes, constraints | Low - after code is stable |

---

## Architecture After Completion

```
src/
├── handlers/
│   └── events.go              # HTTP handlers (thin - orchestrate only)
├── views/                     # Queries stay with views (high cohesion)
│   ├── queries.go
│   ├── all-events_query.go
│   ├── plots_query.go
│   ├── new-event-form_query.go
│   └── events-json_query.go   # NEW
├── services/
│   ├── commands/              # NEW - Command layer
│   │   ├── commands.go        # CommandHandler struct
│   │   └── record-event.go    # RecordEvent command
│   ├── eventstore/
│   │   └── eventstore.go      # Evolves to support event types
│   └── ...
└── services/db/
    └── database.go            # Schema migration
```

---

## Implementation Steps

### Phase 0: Complete Query Layer

#### Step 0.1: Add `GetEventsForJson` Query
- Create `src/views/events-json_query.go` with `EventJson` DTO
- Create `src/views/events-json_query_test.go` with tests
- Migrate `EventsJsonHandler` to use query layer
- Remove `eventstore` import from `EventsJsonHandler`

**Files:**
- New: `views/events-json_query.go`, `views/events-json_query_test.go`
- Modify: `handlers/events.go`, `routes.go`

---

### Phase 1: Database Schema Expansion

#### Step 1.1: Add Schema Migration
- Create migration function in `src/services/db/database.go`
- Add `event_type` column (default 'EventRecorded')
- Add `corrects_event_id` column (nullable)
- Run migration on startup (idempotent)

**Files:**
- Modify: `services/db/database.go`

#### Step 1.2: Update EventStore Interface
- Add `EventType` field to `Event` struct
- Add `CorrectsEventID` field to `Event` struct
- Update `Record()` to accept event type
- Keep backward compatibility

**Files:**
- Modify: `services/eventstore/eventstore.go`

---

### Phase 2: EXPAND - Add Command Layer

#### Step 2.1: Create Command Handler Foundation
- Create `src/services/commands/commands.go`
- `CommandHandler` struct with `eventStore` and `logger` dependencies
- Create `src/services/commands/commands_test.go` with mock

**Files:**
- New: `services/commands/commands.go`, `services/commands/commands_test.go`

#### Step 2.2: Add `RecordEvent` Command (TDD)
- **Red**: Write tests for validation scenarios
- **Green**: Implement `RecordEvent()` method
- Validation: tag format, required fields, value format

**Command signature:**
```go
func (h *CommandHandler) RecordEvent(
    ctx context.Context,
    tag string,
    comment string,
    value string,
    recordedBy string,
) error
```

**Files:**
- New: `services/commands/record-event.go`, `services/commands/record-event_test.go`

---

### Phase 3: MIGRATE - Switch Handlers to Command Layer

#### Step 3.1: Migrate `RecordEventPostHandler`
- Replace direct `eventStore.Record()` with `CommandHandler.RecordEvent()`
- Remove validation logic from handler (move to command)
- Update handler signature

**Files:**
- Modify: `handlers/events.go`, `routes.go`, `main.go`

#### Step 3.2: Remove Unused Dependencies
- Remove `eventstore` import from handlers (all handlers now use commands/queries)
- `eventstore` only imported by `commands` package

**Files:**
- Modify: `handlers/events.go`

---

### Phase 4: CONTRACT - Clean Up

#### Step 4.1: Verify No Regression
- Run `make check`
- Run integration tests
- Verify all handlers use command/query layer

#### Step 4.2: Documentation
- Update AGENTS.md with CQRS architecture
- Document event types and correction flow

---

## Step Summary

| # | Task | New Files | Modified Files | Risk |
|---|------|-----------|----------------|------|
| 0.1 | Add GetEventsForJson query | 2 | 2 | Low |
| 1.1 | Schema migration | 0 | 1 | Low |
| 1.2 | Update EventStore for event types | 0 | 1 | Low |
| 2.1 | CommandHandler foundation | 2 | 0 | Low |
| 2.2 | RecordEvent command (TDD) | 2 | 0 | Medium |
| 3.1 | Migrate RecordEventPostHandler | 0 | 3 | Medium |
| 3.2 | Remove unused dependencies | 0 | 1 | Low |
| 4.1 | Verify and document | 0 | 1 | Low |

---

## Event Type Flow (Future)

### Recording an Event
```
POST /record-event
  → RecordEventCommand
    → EventRecorded (event_type = 'EventRecorded')
```

### Marking an Event as Incorrect (Future)
```
POST /events/{id}/mark-incorrect
  → MarkEventIncorrectCommand
    → EventMarkedAsIncorrect (event_type = 'EventMarkedAsIncorrect', corrects_event_id = {id})
```

### Querying Current State (Future)
```
QueryHandler projects events:
  - EventRecorded → include in projection
  - EventMarkedAsIncorrect → exclude referenced event from projection
```

---

## Rollback Plan

Each step is independently revertible:
1. Revert git commit
2. Run `make check` to confirm stable state
3. Investigate and fix before proceeding

Schema migrations are idempotent (IF NOT EXISTS, ALTER TABLE with defaults).
