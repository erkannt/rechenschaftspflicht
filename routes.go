package main

import (
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/handlers"
	"github.com/erkannt/rechenschaftspflicht/services/authentication"
	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
	"github.com/erkannt/rechenschaftspflicht/services/userstore"
	"github.com/julienschmidt/httprouter"
)

func mustBeLoggedIn(h httprouter.Handle, auth authentication.Auth) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if !auth.IsLoggedIn(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		h(w, r, ps)
	}
}
func addRoutes(
	router *httprouter.Router,
	eventStore eventstore.EventStore,
	userStore userstore.UserStore,
	auth authentication.Auth,
) {
	router.GET("/", handlers.LandingHandler(auth))
	router.POST("/login", handlers.LoginPostHandler(userStore, auth))
	router.GET("/login", handlers.LoginGetHandler(auth))
	router.GET("/check-your-email", handlers.CheckYourEmailHandler)
	router.GET("/record-event", mustBeLoggedIn(handlers.RecordEventFormHandler, auth))
	router.POST("/record-event", mustBeLoggedIn(handlers.RecordEventPostHandler(eventStore, auth), auth))
	router.GET("/all-events", mustBeLoggedIn(handlers.AllEventsHandler(eventStore), auth))
	router.GET("/logout", mustBeLoggedIn(handlers.LogoutHandler, auth))
}
