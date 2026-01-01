package main

import (
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/handlers"
	"github.com/erkannt/rechenschaftspflicht/services"
	"github.com/julienschmidt/httprouter"
)

func mustBeLoggedIn(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		cookie, err := r.Cookie("auth")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if email, err := services.ValidateToken(cookie.Value); err != nil || email == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		h(w, r, ps)
	}
}

func addRoutes(router *httprouter.Router,
	eventStore *services.EventStore) {
	router.GET("/", handlers.LandingHandler)
	router.POST("/login", handlers.LoginPostHandler)
	router.GET("/login", handlers.LoginGetHandler)
	router.GET("/check-your-email", handlers.CheckYourEmailHandler)
	router.GET("/record-event", mustBeLoggedIn(handlers.RecordEventFormHandler))
	router.POST("/record-event", mustBeLoggedIn(handlers.RecordEventPostHandler(eventStore)))
	router.GET("/all-events", mustBeLoggedIn(handlers.AllEventsHandler(eventStore)))
}
