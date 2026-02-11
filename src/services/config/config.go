package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	JWTSecret   string `env:"JWT_SECRET,notEmpty,required"`
	BearerToken string `env:"BEARER_TOKEN,notEmpty,required"`
	SMTPHost    string `env:"SMTP_HOST,notEmpty,required"`
	SMTPPort    string `env:"SMTP_PORT,notEmpty,required"`
	SMTPUser    string `env:"SMTP_USER,required"`
	SMTPPass    string `env:"SMTP_PASS,required"`
	SMTPFrom    string `env:"SMTP_FROM,notEmpty,required"`
	AppOrigin   string `env:"APP_ORIGIN,notEmpty,required"`
	SqlitePath  string `env:"SQLITE_PATH" envDefault:"data/state.db"`
}

func LoadFromEnv() (Config, error) {
	cfg := Config{}

	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return cfg, nil
}
