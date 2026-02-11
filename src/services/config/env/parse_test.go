package env

import (
	"testing"
)

func TestParse(t *testing.T) {
	getenv := func(key string) string {
		m := map[string]string{
			"TEST_VAR": "test_value",
		}
		return m[key]
	}

	type TestConfig struct {
		Field1 string `env:"TEST_VAR"`
		Field2 string `env:"OTHER_VAR"`
	}

	var cfg TestConfig
	err := Parse(getenv, &cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Field1 != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", cfg.Field1)
	}

	if cfg.Field2 != "" {
		t.Errorf("expected '', got '%s'", cfg.Field2)
	}
}

func TestParseSkipsFieldsWithoutTag(t *testing.T) {
	getenv := func(key string) string {
		return "some_value"
	}

	type TestConfig struct {
		FieldWithTag string `env:"SOME_VAR"`
		FieldNoTag   string
	}

	var cfg TestConfig
	err := Parse(getenv, &cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.FieldWithTag != "some_value" {
		t.Errorf("expected 'some_value', got '%s'", cfg.FieldWithTag)
	}

	if cfg.FieldNoTag != "" {
		t.Errorf("expected '', got '%s'", cfg.FieldNoTag)
	}
}
