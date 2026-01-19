package config

import "fmt"

type Config struct {
	JWTSecret string
	SMTPHost  string
	SMTPPort  string
	SMTPUser  string
	SMTPPass  string
	SMTPFrom  string
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{
		JWTSecret: getenv("JWT_SECRET"),
		SMTPHost:  getenv("SMTP_HOST"),
		SMTPPort:  getenv("SMTP_PORT"),
		SMTPUser:  getenv("SMTP_USER"),
		SMTPPass:  getenv("SMTP_PASS"),
		SMTPFrom:  getenv("SMTP_FROM"),
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
