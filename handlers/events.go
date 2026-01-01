package handlers

import (
	"fmt"
	"net/http"
	"time"

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
	views.Layout(views.NewEventForm()).Render(r.Context(), w)
}

func RecordEventPostHandler(store *services.EventStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		tag := r.FormValue("tag")
		comment := r.FormValue("comment")
		value := r.FormValue("value")

		createdAt := time.Now().Format(time.RFC3339)

		event := services.Event{
			Tag:       tag,
			Comment:   comment,
			Value:     value,
			CreatedAt: createdAt,
		}

		// Store the event using the provided EventStore (implementation dependent)
		_ = store // placeholder to avoid unused variable warning

		fmt.Printf("Received: %+v\n", event)

		views.Layout(
			views.NewEventFormWithSuccessBanner(),
		).Render(r.Context(), w)
	}
}
