package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/erkannt/rechenschaftspflicht/services"
	"github.com/erkannt/rechenschaftspflicht/views"
	"github.com/julienschmidt/httprouter"
)

func LandingHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	views.Login().Render(r.Context(), w)
}

func LoginPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		log.Printf("error parsing form: %v", err)
		http.Redirect(w, r, "/check-your-email", http.StatusFound)
		return
	}
	email := r.FormValue("email")
	if email == "" {
		log.Println("email required")
		http.Redirect(w, r, "/check-your-email", http.StatusFound)
		return
	}
	if !services.IsAllowedEmail(email) {
		log.Printf("unauthorized email attempt: %s", email)
		http.Redirect(w, r, "/check-your-email", http.StatusFound)
		return
	}
	token, err := services.GenerateToken(email)
	if err != nil {
		log.Printf("could not generate token for %s: %v", email, err)
		http.Redirect(w, r, "/check-your-email", http.StatusFound)
		return
	}
	if err := services.SendMagicLink(email, token); err != nil {
		log.Printf("could not send email to %s: %v", email, err)
		http.Redirect(w, r, "/check-your-email", http.StatusFound)
		return
	}
	log.Printf("magic login link sent to %s", email)
	http.Redirect(w, r, "/check-your-email", http.StatusFound)
}

func LoginGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	email, err := services.ValidateToken(token)
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

func CheckYourEmailHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	views.CheckYourEmail().Render(r.Context(), w)
}
