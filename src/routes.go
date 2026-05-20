package main

import (
	"embed"
	"log/slog"

	"github.com/erkannt/rechenschaftspflicht/handlers"
	"github.com/erkannt/rechenschaftspflicht/middlewares"
	"github.com/erkannt/rechenschaftspflicht/services/authentication"
	"github.com/erkannt/rechenschaftspflicht/services/config"
	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
	"github.com/erkannt/rechenschaftspflicht/services/userstore"
	"github.com/erkannt/rechenschaftspflicht/views"
	"github.com/julienschmidt/httprouter"
)

//go:embed assets/* assets/**
var embeddedAssets embed.FS

func addRoutes(
	router *httprouter.Router,
	logger *slog.Logger,
	cfg config.Config,
	eventStore eventstore.EventStore,
	userStore userstore.UserStore,
	auth authentication.Auth,
	queryHandler *views.QueryHandler,
) {
	requireLogin := middlewares.MustBeLoggedIn(auth)
	requireBearerToken := middlewares.RequireBearerToken(cfg.BearerToken)

	router.GET("/", handlers.LandingHandler(auth))
	router.POST("/login", handlers.LoginPostHandler(userStore, auth))
	router.GET("/login", handlers.LoginGetHandler(auth))
	router.GET("/check-your-email", handlers.CheckYourEmailHandler)
	router.GET("/record-event", requireLogin(handlers.RecordEventFormHandler(queryHandler, logger)))
	router.POST("/record-event", requireLogin(handlers.RecordEventPostHandler(eventStore, auth, queryHandler, logger)))
	router.GET("/all-events", requireLogin(handlers.AllEventsHandler(queryHandler, logger)))
	router.GET("/events.json", requireLogin(handlers.EventsJsonHandler(queryHandler, logger)))
	router.GET("/plots", requireLogin(handlers.PlotsHandler(queryHandler, logger)))
	router.GET("/logout", requireLogin(handlers.LogoutHandler(auth)))
	router.GET("/oops", handlers.OopsHandler)
	router.POST("/add-user", requireBearerToken(handlers.AddUserHandler(userStore)))

	router.GET("/assets/*filepath", handlers.AssetsHandler(embeddedAssets))
}
