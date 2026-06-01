package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/erkannt/rechenschaftspflicht/services/commands"
	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
	"github.com/erkannt/rechenschaftspflicht/views"
)

// mockEventStore is a test double for eventstore.EventStore
type mockEventStore struct {
	events []eventstore.Event
}

func (m *mockEventStore) RaiseEvent(event eventstore.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventStore) GetAllEvents() ([]eventstore.Event, error) {
	return m.events, nil
}

// mockAuth is a test double for authentication.Auth
type mockAuth struct {
	email string
	err   error
}

func (m *mockAuth) GenerateToken(email string) (string, error)           { return "", nil }
func (m *mockAuth) ValidateToken(tokenStr string) (string, error)        { return "", nil }
func (m *mockAuth) SendMagicLink(toEmail, token string) error            { return nil }
func (m *mockAuth) IsLoggedIn(r *http.Request) bool                      { return true }
func (m *mockAuth) GetLoggedInUserEmail(r *http.Request) (string, error) { return m.email, m.err }
func (m *mockAuth) LoggedIn(token string) http.Cookie                    { return http.Cookie{} }
func (m *mockAuth) LoggedOut() http.Cookie                               { return http.Cookie{} }

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRecordEventPostHandler_ValidationError_RendersFormWithErrors(t *testing.T) {
	mockStore := &mockEventStore{}
	qh := views.NewQueryHandler(mockStore, newTestLogger())
	cmdHandler := commands.NewCommandHandler(mockStore, newTestLogger())
	mockAuth := &mockAuth{email: "user@example.com"}

	handler := RecordEventPostHandler(cmdHandler, mockAuth, qh, newTestLogger())

	// Submit tag with uppercase letter
	form := url.Values{
		"tag":     {"Temperature"},
		"value":   {"10"},
		"comment": {"test comment"},
	}
	req := httptest.NewRequest("POST", "/record-event", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler(w, req, nil)

	// RED: currently returns 500, after fix should return 422
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "There is a problem") {
		t.Error("expected error summary heading")
	}

	if !strings.Contains(body, "lowercase letter") {
		t.Error("expected tag validation error message")
	}

	// Submitted values preserved
	if !strings.Contains(body, `value="Temperature"`) {
		t.Error("expected tag value to be preserved")
	}
	if !strings.Contains(body, "test comment") {
		t.Error("expected comment to be preserved")
	}

	// No event recorded
	if len(mockStore.events) != 0 {
		t.Errorf("expected 0 events recorded, got %d", len(mockStore.events))
	}
}

func TestRecordEventPostHandler_ValidSubmission_RecordsEvent(t *testing.T) {
	mockStore := &mockEventStore{}
	qh := views.NewQueryHandler(mockStore, newTestLogger())
	cmdHandler := commands.NewCommandHandler(mockStore, newTestLogger())
	mockAuth := &mockAuth{email: "user@example.com"}

	handler := RecordEventPostHandler(cmdHandler, mockAuth, qh, newTestLogger())

	form := url.Values{
		"tag":     {"temperature"},
		"value":   {"23.5"},
		"comment": {"cold"},
	}
	req := httptest.NewRequest("POST", "/record-event", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler(w, req, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "Success") {
		t.Error("expected success banner")
	}

	if strings.Contains(body, "There is a problem") {
		t.Error("expected no error summary on success")
	}

	// Event was recorded
	if len(mockStore.events) != 1 {
		t.Fatalf("expected 1 event recorded, got %d", len(mockStore.events))
	}
	if mockStore.events[0].Tag != "temperature" {
		t.Errorf("expected tag 'temperature', got %q", mockStore.events[0].Tag)
	}
	if mockStore.events[0].Value != "23.5" {
		t.Errorf("expected value '23.5', got %q", mockStore.events[0].Value)
	}
}

func TestRecordEventPostHandler_EmptyTag_RendersFormWithError(t *testing.T) {
	mockStore := &mockEventStore{}
	qh := views.NewQueryHandler(mockStore, newTestLogger())
	cmdHandler := commands.NewCommandHandler(mockStore, newTestLogger())
	mockAuth := &mockAuth{email: "user@example.com"}

	handler := RecordEventPostHandler(cmdHandler, mockAuth, qh, newTestLogger())

	form := url.Values{
		"tag":     {""},
		"value":   {""},
		"comment": {""},
	}
	req := httptest.NewRequest("POST", "/record-event", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler(w, req, nil)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "There is a problem") {
		t.Error("expected error summary for empty tag")
	}

	// No event recorded
	if len(mockStore.events) != 0 {
		t.Errorf("expected 0 events recorded, got %d", len(mockStore.events))
	}
}

func TestRecordEventPostHandler_NonNumericValue_RendersFormWithError(t *testing.T) {
	mockStore := &mockEventStore{}
	qh := views.NewQueryHandler(mockStore, newTestLogger())
	cmdHandler := commands.NewCommandHandler(mockStore, newTestLogger())
	mockAuth := &mockAuth{email: "user@example.com"}

	handler := RecordEventPostHandler(cmdHandler, mockAuth, qh, newTestLogger())

	form := url.Values{
		"tag":     {"my-tag"},
		"value":   {"not-a-number"},
		"comment": {""},
	}
	req := httptest.NewRequest("POST", "/record-event", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler(w, req, nil)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "There is a problem") {
		t.Error("expected error summary for non-numeric value")
	}

	if !strings.Contains(body, "valid number") {
		t.Error("expected value validation error message")
	}

	// No event recorded
	if len(mockStore.events) != 0 {
		t.Errorf("expected 0 events recorded, got %d", len(mockStore.events))
	}
}

func TestValidateEventForm_AllValid(t *testing.T) {
	fs := validateEventForm("my-tag", "comment", "23.5")
	if fs.HasErrors() {
		t.Errorf("expected no errors, got %v", fs.Errors)
	}
}

func TestValidateEventForm_EmptyTag(t *testing.T) {
	fs := validateEventForm("", "comment", "10")
	if !fs.HasErrors() {
		t.Fatal("expected errors")
	}
	if fs.ErrorFor("tag") == "" {
		t.Error("expected error for tag field")
	}
	if fs.ErrorFor("value") != "" {
		t.Error("expected no error for valid value")
	}
}

func TestValidateEventForm_UppercaseTag(t *testing.T) {
	fs := validateEventForm("BadTag", "comment", "10")
	if !fs.HasErrors() {
		t.Fatal("expected errors")
	}
	if !strings.Contains(fs.ErrorFor("tag"), "lowercase") {
		t.Error("expected lowercase error message for uppercase tag")
	}
}

func TestValidateEventForm_NonNumericValue(t *testing.T) {
	fs := validateEventForm("my-tag", "comment", "not-a-number")
	if !fs.HasErrors() {
		t.Fatal("expected errors")
	}
	if !strings.Contains(fs.ErrorFor("value"), "valid number") {
		t.Error("expected valid number error message")
	}
}

func TestValidateEventForm_EmptyValueIsAllowed(t *testing.T) {
	fs := validateEventForm("my-tag", "comment", "")
	if fs.HasErrors() {
		t.Errorf("expected no errors for empty optional value, got %v", fs.Errors)
	}
}

func TestValidateEventForm_PreservesSubmittedValues(t *testing.T) {
	fs := validateEventForm("my-tag", "my comment", "42.5")
	if fs.Tag != "my-tag" {
		t.Errorf("expected tag 'my-tag', got %q", fs.Tag)
	}
	if fs.Comment != "my comment" {
		t.Errorf("expected comment 'my comment', got %q", fs.Comment)
	}
	if fs.Value != "42.5" {
		t.Errorf("expected value '42.5', got %q", fs.Value)
	}
}
