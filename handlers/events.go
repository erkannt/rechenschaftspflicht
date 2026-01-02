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
	views.LayoutWithNav(views.NewEventForm()).Render(r.Context(), w)
}

func RecordEventPostHandler(eventStore services.EventStore) httprouter.Handle {
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

		if err := eventStore.Record(event); err != nil {
			fmt.Printf("failed to record event: %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Received: %+v\n", event)

		views.LayoutWithNav(
			views.NewEventFormWithSuccessBanner(),
		).Render(r.Context(), w)
	}
}

func AllEventsHandler(eventStore services.EventStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		events, err := eventStore.GetAll()
		if err != nil {
			fmt.Printf("failed to retrieve events: %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		views.LayoutWithNav(
			views.AllEvents(events),
		).Render(r.Context(), w)
	}
}
