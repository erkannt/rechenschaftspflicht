package authentication

import (
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Auth defines the public contract for the service.
type Auth interface {
	GenerateToken(email string) (string, error)
	ValidateToken(tokenStr string) (string, error)
	SendMagicLink(toEmail, token string) error
	IsLoggedIn(r *http.Request) bool
	GetLoggedInUserEmail(r *http.Request) (string, error)
}

// Config holds configuration required to create a MagicLinks service.
type Config struct {
	JWTSecret string
	SMTPHost  string
	SMTPPort  string
	SMTPUser  string
	SMTPPass  string
	SMTPFrom  string
}

// magicLinksSvc is the concrete implementation holding internal state.
type magicLinksSvc struct {
	jwtSecret []byte
	smtpAuth  smtp.Auth
	smtpFrom  string
	smtpAddr  string
}

// New creates a new MagicLinks service with the supplied configuration.
// It derives sensible defaults from environment variables if fields are empty.
func New(cfg Config) Auth {
	// Apply defaults from environment if not provided.
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = getEnv("JWT_SECRET", "default_secret")
	}
	if cfg.SMTPHost == "" {
		cfg.SMTPHost = getEnv("SMTP_HOST", "localhost")
	}
	if cfg.SMTPPort == "" {
		cfg.SMTPPort = getEnv("SMTP_PORT", "1025")
	}
	if cfg.SMTPFrom == "" {
		cfg.SMTPFrom = getEnv("SMTP_FROM", "no-reply@example.com")
	}

	// Set up SMTP authentication only when a username is supplied.
	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	}

	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)

	return &magicLinksSvc{
		jwtSecret: []byte(cfg.JWTSecret),
		smtpAuth:  auth,
		smtpFrom:  cfg.SMTPFrom,
		smtpAddr:  addr,
	}
}

// GenerateToken creates a JWT containing the email claim that expires in 15 minutes.
func (s *magicLinksSvc) GenerateToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.jwtSecret)
}

// ValidateToken parses and validates the JWT, returning the embedded email if valid.
func (s *magicLinksSvc) ValidateToken(input string) (string, error) {
	token, err := jwt.Parse(input, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if email, ok := claims["email"].(string); ok {
			return email, nil
		}
	}
	return "", fmt.Errorf("email claim missing")
}

// SendMagicLink sends an email containing a login link with the supplied token.
func (s *magicLinksSvc) SendMagicLink(toEmail, token string) error {
	if s.smtpFrom == "" {
		return fmt.Errorf("SMTP configuration incomplete: missing from address")
	}

	link := fmt.Sprintf("http://localhost:8080/login?token=%s", token)
	msg := fmt.Sprintf(
		"From: %s\r\nSubject: Your Magic Login Link\r\n\r\nClick the following link to log in:\n\n%s",
		s.smtpFrom,
		link,
	)

	return smtp.SendMail(s.smtpAddr, s.smtpAuth, s.smtpFrom, []string{toEmail}, []byte(msg))
}

func (s *magicLinksSvc) IsLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("auth")
	if err != nil || cookie.Value == "" {
		return false
	}
	if email, err := s.ValidateToken(cookie.Value); err != nil || email == "" {
		return false
	}
	return true
}

func (s *magicLinksSvc) GetLoggedInUserEmail(r *http.Request) (string, error) {
	cookie, err := r.Cookie("auth")
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", http.ErrNoCookie
	}
	email, err := s.ValidateToken(cookie.Value)
	if err != nil {
		return "", err
	}
	return email, nil
}

func LoggedIn(token string) http.Cookie {
	return http.Cookie{
		Name:     "auth",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   false, // set true when using HTTPS
	}
}

func LoggedOut() http.Cookie {
	return http.Cookie{
		Name:     "auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // Expire immediately
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // set true when using HTTPS
	}
}
