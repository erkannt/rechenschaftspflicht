package handlers

import (
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/services"
	"github.com/erkannt/rechenschaftspflicht/views"
	"github.com/julienschmidt/httprouter"
)

func RecordEventFormHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cookie, err := r.Cookie("auth")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if email, err := services.ValidateToken(cookie.Value); err != nil || email == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	views.Layout(views.LogNewEvent()).Render(r.Context(), w)
}
