package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/julienschmidt/httprouter"
)

// ---------------------------------------------------------------------
// Configuration & helpers
// ---------------------------------------------------------------------

var (
	allowedEmails = []string{
		"foo@example.com",
		"alice@example.com",
		"bob@example.com",
	}
	jwtSecret = []byte(getEnv("JWT_SECRET", "default_secret"))
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func isAllowedEmail(email string) bool {
	for _, e := range allowedEmails {
		if strings.EqualFold(e, email) {
			return true
		}
	}
	return false
}

// generate a JWT token that expires in 15 minutes
func generateToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(jwtSecret)
}

// validate a JWT token and return the contained email address
func validateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
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

// send a magic login link via SMTP
func sendMagicLink(toEmail, token string) error {
	smtpHost := getEnv("SMTP_HOST", "")
	smtpPort := getEnv("SMTP_PORT", "587")
	smtpUser := getEnv("SMTP_USER", "")
	smtpPass := getEnv("SMTP_PASS", "")
	smtpFrom := getEnv("SMTP_FROM", "")

	if smtpHost == "" || smtpFrom == "" {
		return fmt.Errorf("SMTP configuration incomplete")
	}

	link := fmt.Sprintf("http://localhost:8080/login?token=%s", token)
	msg := fmt.Sprintf("Subject: Your Magic Login Link\r\n\r\nClick the following link to log in:\n\n%s", link)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	return smtp.SendMail(addr, auth, smtpFrom, []string{toEmail}, []byte(msg))
}

// ---------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------

// GET "/" – simple landing page with a form
func rootHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	login().Render(r.Context(), w)
}

// POST "/login" – receive email, validate, send magic link
func loginPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "email required", http.StatusBadRequest)
		return
	}
	if !isAllowedEmail(email) {
		http.Error(w, "email not authorized", http.StatusForbidden)
		return
	}
	token, err := generateToken(email)
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}
	if err := sendMagicLink(email, token); err != nil {
		http.Error(w, "could not send email", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "Magic login link sent (check your email).")
}

// GET "/login" – validate token and redirect
func loginGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	email, err := validateToken(token)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Set a short‑lived auth cookie
	cookie := &http.Cookie{
		Name:     "auth",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   false, // set true when using HTTPS
	}
	http.SetCookie(w, cookie)

	// Optionally log the successful login
	log.Printf("User %s logged in via magic link", email)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// GET "/dashboard" – protected resource
func dashboardHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cookie, err := r.Cookie("auth")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusTooManyRequests) // 429
		return
	}
	if email, err := validateToken(cookie.Value); err != nil || email == "" {
		w.WriteHeader(http.StatusTooManyRequests) // 429
		return
	}
	fmt.Fprint(w, "success")
}

// ---------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------

func main() {
	router := httprouter.New()
	router.GET("/", rootHandler)
	router.POST("/login", loginPostHandler)
	router.GET("/login", loginGetHandler)
	router.GET("/dashboard", dashboardHandler)

	srv := &http.Server{Addr: ":8080", Handler: router}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Println("Server is listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}

	log.Println("Server stopped")
}
