package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/erkannt/rechenschaftspflicht/services/authentication"
	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
	"github.com/erkannt/rechenschaftspflicht/views"
	"github.com/julienschmidt/httprouter"
)

func RecordEventFormHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := views.LayoutWithNav(views.NewEventForm()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error rendering layout: %v", err)
		return
	}
}

func RecordEventPostHandler(eventStore eventstore.EventStore, auth authentication.Auth) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		tag := r.FormValue("tag")
		comment := r.FormValue("comment")
		value := r.FormValue("value")

		recordedAt := time.Now().Format(time.RFC3339)
		recordedBy, _ := auth.GetLoggedInUserEmail(r)

		event := eventstore.Event{
			Tag:        tag,
			Comment:    comment,
			Value:      value,
			RecordedAt: recordedAt,
			RecordedBy: recordedBy,
		}

		if err := eventStore.Record(event); err != nil {
			fmt.Printf("failed to record event: %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Received: %+v\n", event)

		err := views.LayoutWithNav(views.NewEventFormWithSuccessBanner()).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error rendering layout: %v", err)
			return
		}
	}
}

func AllEventsHandler(eventStore eventstore.EventStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		events, err := eventStore.GetAll()
		if err != nil {
			fmt.Printf("failed to retrieve events: %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = views.LayoutWithNav(views.AllEvents(events)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error rendering layout: %v", err)
			return
		}
	}
}

func PlotsHandler(eventStore eventstore.EventStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		events, err := eventStore.GetAll()
		if err != nil {
			fmt.Printf("failed to retrieve events: %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = views.LayoutWithNav(views.Plots(events)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error rendering layout: %v", err)
			return
		}
	}
}
