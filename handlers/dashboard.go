package handlers

import (
	"fmt"
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/services"
	"github.com/julienschmidt/httprouter"
)

// GET "/dashboard" – protected resource
func DashboardHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cookie, err := r.Cookie("auth")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusTooManyRequests) // 429
		return
	}
	if email, err := services.ValidateToken(cookie.Value); err != nil || email == "" {
		w.WriteHeader(http.StatusTooManyRequests) // 429
		return
	}
	fmt.Fprint(w, "success")
}
