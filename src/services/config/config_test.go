package config

import (
	"testing"
)

func TestConfigValid(t *testing.T) {
	cfg := Config{
		JWTSecret:   "secret",
		BearerToken: "token",
		SMTPHost:    "localhost",
		SMTPPort:    "587",
		SMTPFrom:    "from@test.com",
		AppOrigin:   "http://localhost:3000",
	}

	problems := cfg.Valid()
	if len(problems) > 0 {
		t.Errorf("expected no problems, got: %v", problems)
	}
}

func TestConfigValidMissingFields(t *testing.T) {
	cfg := Config{}

	problems := cfg.Valid()
	if len(problems) != 6 {
		t.Errorf("expected 6 problems, got %d: %v", len(problems), problems)
	}

	if _, ok := problems["JWTSecret"]; !ok {
		t.Error("expected JWTSecret problem")
	}
	if _, ok := problems["BearerToken"]; !ok {
		t.Error("expected BearerToken problem")
	}
	if _, ok := problems["SMTPHost"]; !ok {
		t.Error("expected SMTPHost problem")
	}
	if _, ok := problems["SMTPPort"]; !ok {
		t.Error("expected SMTPPort problem")
	}
	if _, ok := problems["SMTPFrom"]; !ok {
		t.Error("expected SMTPFrom problem")
	}
	if _, ok := problems["AppOrigin"]; !ok {
		t.Error("expected AppOrigin problem")
	}
}

func TestProblemsToError(t *testing.T) {
	problems := Problems{
		"JWTSecret": "JWT_SECRET is required",
		"SMTPHost":  "SMTP_HOST is required",
	}

	err := problemsToError(problems)
	if err == nil {
		t.Error("expected error, got nil")
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("expected non-empty error message")
	}
	if len(errStr) == 0 {
		t.Error("expected error message to contain content")
	}
}

func TestProblemsToErrorNil(t *testing.T) {
	err := problemsToError(Problems{})
	if err != nil {
		t.Errorf("expected nil error, got '%s'", err.Error())
	}
}
