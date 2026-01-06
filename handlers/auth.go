package handlers

import (
	"log"
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/services/authcookie"
	"github.com/erkannt/rechenschaftspflicht/services/magiclinks"
	"github.com/erkannt/rechenschaftspflicht/services/userstore"
	"github.com/erkannt/rechenschaftspflicht/views"
	"github.com/julienschmidt/httprouter"
)

func LandingHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if authcookie.IsLoggedIn(r) {
		cookie, _ := r.Cookie("auth")
		email, _ := magiclinks.ValidateToken(cookie.Value)
		log.Printf("User %s already logged in, redirecting to /record-event", email)
		http.Redirect(w, r, "/record-event", http.StatusFound)
		return
	}
	err := views.LayoutBare(views.Login()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error rendering layout: %v", err)
		return
	}
}

func LoginPostHandler(userStore userstore.UserStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

		exists, err := userStore.IsUser(email)
		if err != nil {
			log.Printf("error checking user %s: %v", email, err)
			http.Redirect(w, r, "/check-your-email", http.StatusFound)
			return
		}
		if !exists {
			log.Printf("unauthorized email attempt: %s", email)
			http.Redirect(w, r, "/check-your-email", http.StatusFound)
			return
		}

		token, err := magiclinks.GenerateToken(email)
		if err != nil {
			log.Printf("could not generate token for %s: %v", email, err)
			http.Redirect(w, r, "/check-your-email", http.StatusFound)
			return
		}
		if err := magiclinks.SendMagicLink(email, token); err != nil {
			log.Printf("could not send email to %s: %v", email, err)
			http.Redirect(w, r, "/check-your-email", http.StatusFound)
			return
		}
		log.Printf("magic login link sent to %s", email)

		http.Redirect(w, r, "/check-your-email", http.StatusFound)
	}
}

func LoginGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	email, err := magiclinks.ValidateToken(token)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	cookie := authcookie.LoggedIn(token)
	http.SetCookie(w, &cookie)

	log.Printf("User %s logged in via magic link", email)
	http.Redirect(w, r, "/record-event", http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cookie := authcookie.LoggedOut()
	http.SetCookie(w, &cookie)

	log.Println("User logged out")
	http.Redirect(w, r, "/", http.StatusFound)
}

func CheckYourEmailHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := views.LayoutBare(views.CheckYourEmail()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error rendering layout: %v", err)
		return
	}
}
