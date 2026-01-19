package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	JWTSecret string `env:"JWT_SECRET"`
	SMTPHost  string `env:"SMTP_HOST"`
	SMTPPort  string `env:"SMTP_PORT"`
	SMTPUser  string `env:"SMTP_USER"`
	SMTPPass  string `env:"SMTP_PASS"`
	SMTPFrom  string `env:"SMTP_FROM"`
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is not set")
	}
	if cfg.SMTPHost == "" {
		return Config{}, fmt.Errorf("SMTP_HOST is not set")
	}
	if cfg.SMTPPort == "" {
		return Config{}, fmt.Errorf("SMTP_PORT is not set")
	}
	if cfg.SMTPUser == "" {
		return Config{}, fmt.Errorf("SMTP_USER is not set")
	}
	if cfg.SMTPPass == "" {
		return Config{}, fmt.Errorf("SMTP_PASS is not set")
	}
	if cfg.SMTPFrom == "" {
		return Config{}, fmt.Errorf("SMTP_FROM is not set")
	}

	return cfg, nil
}
