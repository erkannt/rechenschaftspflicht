package main

import (
	"github.com/erkannt/rechenschaftspflicht/handlers"
	"github.com/julienschmidt/httprouter"
)

func addRoutes(router *httprouter.Router) {
	router.GET("/", handlers.LandingHandler)
	router.POST("/login", handlers.LoginPostHandler)
	router.GET("/login", handlers.LoginGetHandler)
	router.GET("/check-your-email", handlers.CheckYourEmailHandler)
	router.GET("/dashboard", handlers.LogEventFormHandler)
}
