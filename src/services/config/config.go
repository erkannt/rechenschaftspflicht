package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	JWTSecret string `env:"JWT_SECRET,notEmpty"`
	SMTPHost  string `env:"SMTP_HOST,notEmpty"`
	SMTPPort  string `env:"SMTP_PORT,notEmpty"`
	SMTPUser  string `env:"SMTP_USER,required"`
	SMTPPass  string `env:"SMTP_PASS,required"`
	SMTPFrom  string `env:"SMTP_FROM,notEmpty"`
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{}

	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return cfg, nil
}
