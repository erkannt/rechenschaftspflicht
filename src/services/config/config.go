package config

import (
	"fmt"
	"strings"

	"github.com/erkannt/rechenschaftspflicht/services/config/env"
)

type Problems map[string]string

type Config struct {
	JWTSecret   string `env:"JWT_SECRET"`
	BearerToken string `env:"BEARER_TOKEN"`
	SMTPHost    string `env:"SMTP_HOST"`
	SMTPPort    string `env:"SMTP_PORT"`
	SMTPUser    string `env:"SMTP_USER"`
	SMTPPass    string `env:"SMTP_PASS"`
	SMTPFrom    string `env:"SMTP_FROM"`
	AppOrigin   string `env:"APP_ORIGIN"`
	SqlitePath  string `env:"SQLITE_PATH"`
}

var defaultConfig = Config{
	SqlitePath: "data/state.db",
}

func (c Config) Valid() Problems {
	problems := Problems{}

	if c.JWTSecret == "" {
		problems["JWTSecret"] = "JWT_SECRET is required"
	}
	if c.BearerToken == "" {
		problems["BearerToken"] = "BEARER_TOKEN is required"
	}
	if c.SMTPHost == "" {
		problems["SMTPHost"] = "SMTP_HOST is required"
	}
	if c.SMTPPort == "" {
		problems["SMTPPort"] = "SMTP_PORT is required"
	}
	if c.SMTPFrom == "" {
		problems["SMTPFrom"] = "SMTP_FROM is required"
	}
	if c.AppOrigin == "" {
		problems["AppOrigin"] = "APP_ORIGIN is required"
	}

	return problems
}

func problemsToError(problems Problems) error {
	if len(problems) == 0 {
		return nil
	}

	var msgs []string
	for field, msg := range problems {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return fmt.Errorf("validation failed: %s", strings.Join(msgs, ", "))
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	cfg := defaultConfig

	if err := env.Parse(getenv, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse env: %w", err)
	}

	if err := problemsToError(cfg.Valid()); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
