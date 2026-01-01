package main

import (
	"github.com/erkannt/rechenschaftspflicht/handlers"
	"github.com/erkannt/rechenschaftspflicht/services"
	"github.com/julienschmidt/httprouter"
)

func addRoutes(router *httprouter.Router,
	eventStore *services.EventStore) {
	router.GET("/", handlers.LandingHandler)
	router.POST("/login", handlers.LoginPostHandler)
	router.GET("/login", handlers.LoginGetHandler)
	router.GET("/check-your-email", handlers.CheckYourEmailHandler)
	router.GET("/record-event", handlers.RecordEventFormHandler)
	router.POST("/record-event", handlers.RecordEventPostHandler(eventStore))
	router.GET("/all-events", handlers.AllEventsHandler(eventStore))
}
