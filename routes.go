package main

import (
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/handlers"
	"github.com/erkannt/rechenschaftspflicht/services"
	"github.com/julienschmidt/httprouter"
)

func mustBeLoggedIn(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !handlers.IsLoggedIn(r) {
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
	router.GET("/logout", mustBeLoggedIn(handlers.LogoutHandler))
}
