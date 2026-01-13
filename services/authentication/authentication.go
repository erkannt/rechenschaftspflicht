package authentication

import (
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "default_secret"))

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GenerateToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(jwtSecret)
}

func ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
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

func SendMagicLink(toEmail, token string) error {
	smtpHost := getEnv("SMTP_HOST", "localhost")
	smtpPort := getEnv("SMTP_PORT", "1025")
	smtpUser := getEnv("SMTP_USER", "")
	smtpPass := getEnv("SMTP_PASS", "")
	smtpFrom := getEnv("SMTP_FROM", "no-reply@example.com")

	if smtpHost == "" || smtpFrom == "" {
		return fmt.Errorf("SMTP configuration incomplete")
	}

	link := fmt.Sprintf("http://localhost:8080/login?token=%s", token)
	msg := fmt.Sprintf("From: %s\r\nSubject: Your Magic Login Link\r\n\r\nClick the following link to log in:\n\n%s",
		smtpFrom, link)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	// MailCatcher does not support AUTH, so only use authentication when credentials are provided.
	var auth smtp.Auth
	if smtpUser != "" {
		auth = smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	}

	return smtp.SendMail(addr, auth, smtpFrom, []string{toEmail}, []byte(msg))
}


func IsLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("auth")
	if err != nil || cookie.Value == "" {
		return false
	}
	if email, err := ValidateToken(cookie.Value); err != nil || email == "" {
		return false
	}
	return true
}

func GetLoggedInUserEmail(r *http.Request) (string, error) {
	cookie, err := r.Cookie("auth")
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", http.ErrNoCookie
	}
	email, err := ValidateToken(cookie.Value)
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
