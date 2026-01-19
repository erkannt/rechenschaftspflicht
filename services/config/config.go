package config

type Config struct {
	JWTSecret string
	SMTPHost  string
	SMTPPort  string
	SMTPUser  string
	SMTPPass  string
	SMTPFrom  string
}
