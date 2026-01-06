package authcookie

import (
	"net/http"
	"time"

	"github.com/erkannt/rechenschaftspflicht/services/magiclinks"
)

func IsLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("auth")
	if err != nil || cookie.Value == "" {
		return false
	}
	if email, err := magiclinks.ValidateToken(cookie.Value); err != nil || email == "" {
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
	email, err := magiclinks.ValidateToken(cookie.Value)
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
