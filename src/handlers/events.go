package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/erkannt/rechenschaftspflicht/services/authentication"
	"github.com/erkannt/rechenschaftspflicht/services/commands"
	"github.com/erkannt/rechenschaftspflicht/views"
	"github.com/julienschmidt/httprouter"
)

func RecordEventFormHandler(queries *views.QueryHandler, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		tags, err := queries.GetTagSuggestions()
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to retrieve tags", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = views.LayoutWithNav(views.NewEventForm(tags, views.FormState{})).Render(r.Context(), w)
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to render event form", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func RecordEventPostHandler(cmdHandler *commands.CommandHandler, auth authentication.Auth, queries *views.QueryHandler, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		tag := r.FormValue("tag")
		comment := r.FormValue("comment")
		value := r.FormValue("value")

		// Build form state with submitted values for re-rendering on error
		formState := views.FormState{
			Tag:     tag,
			Value:   value,
			Comment: comment,
			Errors:  make(map[string]string),
		}

		// Validate tag
		if tag == "" {
			formState.Errors["tag"] = "Enter a tag"
		} else if !commands.TagPattern.MatchString(tag) {
			formState.Errors["tag"] = "Tag must start with a lowercase letter and contain only lowercase letters and hyphens"
		}

		// Validate value (if provided, must be numeric)
		if value != "" {
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				formState.Errors["value"] = "Value must be a valid number"
			}
		}

		// If validation fails, re-render the form with errors
		if formState.HasErrors() {
			tags, err := queries.GetTagSuggestions()
			if err != nil {
				logger.ErrorContext(r.Context(), "failed to retrieve tags for error form", slog.Any("error", err))
				tags = nil
			}
			w.WriteHeader(http.StatusUnprocessableEntity)
			err = views.LayoutWithNav(views.NewEventForm(tags, formState)).Render(r.Context(), w)
			if err != nil {
				logger.ErrorContext(r.Context(), "failed to render error form", slog.Any("error", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		recordedBy, _ := auth.GetLoggedInUserEmail(r)

		if err := cmdHandler.RecordEvent(r.Context(), tag, comment, value, recordedBy); err != nil {
			logger.ErrorContext(r.Context(), "failed to record event", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

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

func AllEventsHandler(queries *views.QueryHandler, auth authentication.Auth, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		currentUserEmail, _ := auth.GetLoggedInUserEmail(r)
		events, err := queries.GetAllEventsForListWithCurrentUser(currentUserEmail)
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

func EventsJsonHandler(queries *views.QueryHandler, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		events, err := queries.GetEventsForJson()
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to retrieve events", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
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

func MarkEventIncorrectPostHandler(cmdHandler *commands.CommandHandler, auth authentication.Auth, logger *slog.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		sequenceStr := r.FormValue("sequence")
		if sequenceStr == "" {
			http.Error(w, "sequence is required", http.StatusBadRequest)
			return
		}

		sequence, err := strconv.Atoi(sequenceStr)
		if err != nil {
			http.Error(w, "invalid sequence number", http.StatusBadRequest)
			return
		}

		markedBy, err := auth.GetLoggedInUserEmail(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if err := cmdHandler.MarkEventAsIncorrect(r.Context(), sequence, markedBy); err != nil {
			logger.ErrorContext(r.Context(), "failed to mark event as incorrect", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Redirect back to all events page
	http.Redirect(w, r, "/all-events", http.StatusSeeOther)
	}
}
