package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/erkannt/rechenschaftspflicht/services/authentication"
	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
	"github.com/erkannt/rechenschaftspflicht/views"
	"github.com/julienschmidt/httprouter"
)

type EventResponse struct {
	Tag        string  `json:"tag"`
	Comment    string  `json:"comment"`
	Value      string  `json:"value"`
	ValueNum   float64 `json:"valueNum"`
	RecordedAt string  `json:"recordedAt"`
	RecordedBy string  `json:"recordedBy"`
}

func RecordEventFormHandler(queries *views.QueryHandler, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		tags, err := queries.GetTagSuggestions()
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to retrieve tags", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = views.LayoutWithNav(views.NewEventForm(tags)).Render(r.Context(), w)
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to render event form", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func RecordEventPostHandler(eventStore eventstore.EventStore, auth authentication.Auth, queries *views.QueryHandler, logger *slog.Logger) httprouter.Handle {
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
			logger.ErrorContext(r.Context(), "failed to record event", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		logger.DebugContext(r.Context(), "event recorded",
			slog.String("tag", event.Tag),
			slog.String("recordedBy", event.RecordedBy),
		)

		tags, err := queries.GetTagSuggestions()
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to retrieve tags for success form", slog.Any("error", err))
		}

		err = views.LayoutWithNav(views.NewEventFormWithSuccessBanner(tags)).Render(r.Context(), w)
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to render success form", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func AllEventsHandler(queries *views.QueryHandler, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		events, err := queries.GetAllEventsForList()
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to retrieve events", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = views.LayoutWithNav(views.AllEvents(events)).Render(r.Context(), w)
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to render all events", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func EventsJsonHandler(eventStore eventstore.EventStore, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		events, err := eventStore.GetAll()
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to retrieve events", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		var eventResponses []EventResponse
		for _, event := range events {
			if event.Value == "" {
				continue
			}

			valueNum, err := strconv.ParseFloat(event.Value, 64)
			if err != nil {
				continue
			}

			eventResponse := EventResponse{
				Tag:        event.Tag,
				Comment:    event.Comment,
				Value:      event.Value,
				ValueNum:   valueNum,
				RecordedAt: event.RecordedAt,
				RecordedBy: event.RecordedBy,
			}
			eventResponses = append(eventResponses, eventResponse)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(eventResponses); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode events to JSON", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func PlotsHandler(queries *views.QueryHandler, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		events, err := queries.GetEventsForPlots()
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to retrieve events for plots", slog.Any("error", err))
			http.Error(w, "Failed to retrieve events", http.StatusInternalServerError)
			return
		}
		err = views.LayoutWithNav(views.Plots(events)).Render(r.Context(), w)
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to render plots", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
