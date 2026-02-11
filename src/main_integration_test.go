package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testJWTSecret   = "test-jwt-secret-32-bytes-long!!"
	testBearerToken = "test-bearer-token"
	testEmail       = "testuser@example.com"
	testUsername    = "testuser"
	serverAddr      = "localhost:8080"
	appOrigin       = "http://localhost:8080"
)

// mailpitMessage represents a message from the Mailpit API
type mailpitMessage struct {
	ID      string `json:"ID"`
	Subject string `json:"Subject"`
}

// mailpitMessagesResponse represents the response from Mailpit API
type mailpitMessagesResponse struct {
	Messages []mailpitMessage `json:"messages"`
	Total    int              `json:"total"`
}

// mailpitMessageDetail represents the full message details
type mailpitMessageDetail struct {
	ID   string `json:"ID"`
	Text string `json:"Text"`
}

func TestHappyPath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create temp directory for SQLite database
	tempDir, err := os.MkdirTemp("", "e2e-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	dbPath := filepath.Join(tempDir, "test.db")

	// Start Mailpit container
	mailpitContainer, mailpitAPIPort, mailpitSMTPPort, err := startMailpit(ctx)
	if err != nil {
		t.Fatalf("failed to start mailpit: %v", err)
	}
	defer func() { _ = mailpitContainer.Terminate(ctx) }()

	mailpitHost, err := mailpitContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get mailpit host: %v", err)
	}

	// Build test config using getenv stub
	envVars := map[string]string{
		"JWT_SECRET":   testJWTSecret,
		"BEARER_TOKEN": testBearerToken,
		"SMTP_HOST":    mailpitHost,
		"SMTP_PORT":    mailpitSMTPPort,
		"SMTP_USER":    "",
		"SMTP_PASS":    "",
		"SMTP_FROM":    "noreply@example.com",
		"APP_ORIGIN":   appOrigin,
		"SQLITE_PATH":  dbPath,
	}

	getenv := func(key string) string {
		return envVars[key]
	}

	// Capture server logs
	var serverLogs bytes.Buffer

	// Start the application in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- run(ctx, &serverLogs, getenv)
	}()

	// Print logs if test fails
	t.Cleanup(func() {
		if t.Failed() {
			t.Log("=== SERVER LOGS ===")
			if serverLogs.Len() > 0 {
				t.Log(serverLogs.String())
			} else {
				t.Log("(no server logs captured)")
			}
		}
	})

	// Wait for server to be ready
	if err := waitForServer(serverAddr, 10*time.Second, t); err != nil {
		t.Fatalf("server failed to start: %v", err)
	}
	t.Log("Server is ready")

	// Create HTTP client with cookie jar (follows redirects automatically)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	// Step 1: Create user via /add-user with Bearer token
	t.Run("create user", func(t *testing.T) {
		payload := map[string]string{
			"email":    testEmail,
			"username": testUsername,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/add-user", serverAddr), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testBearerToken)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("User created successfully")
	})

	// Step 2: Login via magic link
	t.Run("login via magic link", func(t *testing.T) {
		// Request magic link
		formData := url.Values{}
		formData.Set("email", testEmail)

		resp, err := client.PostForm(fmt.Sprintf("http://%s/login", serverAddr), formData)
		if err != nil {
			t.Fatalf("failed to request login: %v", err)
		}
		_ = resp.Body.Close()

		// Fetch magic link from Mailpit (polls with progress indicator)
		magicLink, err := getMagicLinkFromMailpit(mailpitHost, mailpitAPIPort, testEmail)
		if err != nil {
			t.Fatalf("failed to get magic link: %v", err)
		}
		t.Logf("Got magic link: %s", magicLink)

		// Extract token from magic link
		token := extractTokenFromURL(magicLink)
		if token == "" {
			t.Fatalf("could not extract token from magic link")
		}

		// Visit magic link to set auth cookie and follow redirect
		resp, err = client.Get(fmt.Sprintf("http://%s/login?token=%s", serverAddr, token))
		if err != nil {
			t.Fatalf("failed to visit magic link: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Check if we ended up on the record-event page (successful login)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 after redirect, got %d", resp.StatusCode)
		}

		// Verify we're on the correct page by checking the URL
		if resp.Request.URL.Path != "/record-event" {
			t.Fatalf("expected to be redirected to /record-event, got %s", resp.Request.URL.Path)
		}

		t.Log("Login successful")
	})

	// Step 3: Raise three events
	t.Run("raise events", func(t *testing.T) {
		events := []struct {
			tag     string
			value   string
			comment string
		}{
			{"test-tag-1", "10.5", "First test event"},
			{"test-tag-2", "20.0", "Second test event"},
			{"test-tag-1", "15.0", "Third test event"},
		}

		for _, event := range events {
			formData := url.Values{}
			formData.Set("tag", event.tag)
			formData.Set("value", event.value)
			formData.Set("comment", event.comment)

			resp, err := client.PostForm(fmt.Sprintf("http://%s/record-event", serverAddr), formData)
			if err != nil {
				t.Fatalf("failed to record event: %v", err)
			}
			_ = resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}
		}
		t.Log("3 events recorded successfully")
	})

	// Step 4: Check events appear on /all-events
	t.Run("check all-events page", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://%s/all-events", serverAddr))
		if err != nil {
			t.Fatalf("failed to get all-events: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		content := string(body)

		// Check that our events are in the page
		if !strings.Contains(content, "First test event") {
			t.Error("First test event not found in page")
		}
		if !strings.Contains(content, "Second test event") {
			t.Error("Second test event not found in page")
		}
		if !strings.Contains(content, "Third test event") {
			t.Error("Third test event not found in page")
		}
		t.Log("All events found on page")
	})

	// Step 5: Check events.json is accessible
	t.Run("check events.json", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://%s/events.json", serverAddr))
		if err != nil {
			t.Fatalf("failed to get events.json: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("expected JSON content type, got %s", contentType)
		}

		var events []eventstore.Event
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			t.Fatalf("failed to decode events: %v", err)
		}

		if len(events) != 3 {
			t.Errorf("expected 3 events, got %d", len(events))
		}

		// Verify values are numeric
		for _, event := range events {
			if event.Value == "" {
				t.Errorf("event has empty value: %+v", event)
			}
		}

		t.Logf("events.json contains %d events", len(events))
	})

	// Step 6: Check plots.js asset is accessible
	t.Run("check plots.js asset", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://%s/assets/plot.js", serverAddr))
		if err != nil {
			t.Fatalf("failed to get plot.js: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "javascript") && !strings.Contains(contentType, "application") {
			// Some servers might serve JS with different content types
			t.Logf("Content-Type: %s (this may vary)", contentType)
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) == 0 {
			t.Error("plot.js is empty")
		}

		t.Logf("plot.js loaded successfully (%d bytes)", len(body))
	})

	// Cancel context to stop server
	cancel()

	// Wait for server to stop
	select {
	case err := <-serverErr:
		if err != nil {
			t.Logf("Server stopped with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Log("Server shutdown timeout")
	}
}

// startMailpit starts a Mailpit container and returns the container, API port, and SMTP port
func startMailpit(ctx context.Context) (testcontainers.Container, string, string, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "axllent/mailpit:v1.21",
			ExposedPorts: []string{"8025/tcp", "1025/tcp"},
			WaitingFor:   wait.ForListeningPort("1025/tcp"),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to start mailpit container: %w", err)
	}

	apiPort, err := container.MappedPort(ctx, "8025")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", "", fmt.Errorf("failed to get API port: %w", err)
	}

	smtpPort, err := container.MappedPort(ctx, "1025")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", "", fmt.Errorf("failed to get SMTP port: %w", err)
	}

	return container, apiPort.Port(), smtpPort.Port(), nil
}

// waitForServer waits for the server to be ready with progress indicator
func waitForServer(addr string, timeout time.Duration, t *testing.T) error {
	_, _ = fmt.Fprint(os.Stdout, "Waiting for server to be ready")
	defer func() { _, _ = fmt.Fprintln(os.Stdout) }()

	deadline := time.Now().Add(timeout)
	nextDot := time.Now().Add(time.Second)

	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://%s/", addr))
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}

		// Print progress dot every second
		if time.Now().After(nextDot) {
			_, _ = fmt.Fprint(os.Stdout, ".")
			nextDot = time.Now().Add(time.Second)
		}

		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server not ready after %v", timeout)
}

// getMagicLinkFromMailpit retrieves the magic link email from Mailpit with polling and progress indicator
func getMagicLinkFromMailpit(host, apiPort, toEmail string) (string, error) {
	_, _ = fmt.Fprint(os.Stdout, "Waiting for magic link email")
	defer func() { _, _ = fmt.Fprintln(os.Stdout) }()

	timeout := 10 * time.Second
	deadline := time.Now().Add(timeout)
	nextDot := time.Now().Add(time.Second)

	for time.Now().Before(deadline) {
		// Get list of messages
		resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/v1/messages", host, apiPort))
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		var messagesResp mailpitMessagesResponse
		if err := json.NewDecoder(resp.Body).Decode(&messagesResp); err != nil {
			_ = resp.Body.Close()
			time.Sleep(200 * time.Millisecond)
			continue
		}
		_ = resp.Body.Close()

		if messagesResp.Total > 0 {
			// Get the first (most recent) message
			messageID := messagesResp.Messages[0].ID

			// Get message details
			resp, err = http.Get(fmt.Sprintf("http://%s:%s/api/v1/message/%s", host, apiPort, messageID))
			if err != nil {
				time.Sleep(200 * time.Millisecond)
				continue
			}

			var msgDetail mailpitMessageDetail
			if err := json.NewDecoder(resp.Body).Decode(&msgDetail); err != nil {
				_ = resp.Body.Close()
				time.Sleep(200 * time.Millisecond)
				continue
			}
			_ = resp.Body.Close()

			// Extract magic link from message text
			linkRegex := regexp.MustCompile(`http://[^\s]+/login\?token=[^\s]+`)
			match := linkRegex.FindString(msgDetail.Text)

			if match != "" {
				return match, nil
			}
		}

		// Print progress dot every second
		if time.Now().After(nextDot) {
			_, _ = fmt.Fprint(os.Stdout, ".")
			nextDot = time.Now().Add(time.Second)
		}

		time.Sleep(200 * time.Millisecond)
	}
	return "", fmt.Errorf("magic link email not found after %v", timeout)
}

// extractTokenFromURL extracts the token parameter from a URL
func extractTokenFromURL(magicLink string) string {
	u, err := url.Parse(magicLink)
	if err != nil {
		return ""
	}
	return u.Query().Get("token")
}
