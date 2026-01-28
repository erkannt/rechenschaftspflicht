package main

import (
	"net/http"

	"github.com/erkannt/rechenschaftspflicht/handlers"
	"github.com/erkannt/rechenschaftspflicht/services/authentication"
	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
	"github.com/erkannt/rechenschaftspflicht/services/userstore"
	"github.com/julienschmidt/httprouter"
)

func mustBeLoggedIn(auth authentication.Auth) func(httprouter.Handle) httprouter.Handle {
	return func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			if !auth.IsLoggedIn(r) {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			h(w, r, ps)
		}
	}
}

func addRoutes(
	router *httprouter.Router,
	eventStore eventstore.EventStore,
	userStore userstore.UserStore,
	auth authentication.Auth,
) {
	requireLogin := mustBeLoggedIn(auth)

	router.GET("/", handlers.LandingHandler(auth))
	router.POST("/login", handlers.LoginPostHandler(userStore, auth))
	router.GET("/login", handlers.LoginGetHandler(auth))
	router.GET("/check-your-email", handlers.CheckYourEmailHandler)
	router.GET("/record-event", requireLogin(handlers.RecordEventFormHandler))
	router.POST("/record-event", requireLogin(handlers.RecordEventPostHandler(eventStore, auth)))
	router.GET("/all-events", requireLogin(handlers.AllEventsHandler(eventStore)))
	router.GET("/logout", requireLogin(handlers.LogoutHandler))
}
