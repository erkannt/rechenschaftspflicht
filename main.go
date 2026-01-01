package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erkannt/rechenschaftspflicht/handlers"
	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.GET("/", handlers.LandingHandler)
	router.POST("/login", handlers.LoginPostHandler)
	router.GET("/login", handlers.LoginGetHandler)
	router.GET("/check-your-email", handlers.CheckYourEmailHandler)
	router.GET("/dashboard", handlers.DashboardHandler)

	srv := &http.Server{Addr: ":8080", Handler: router}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Println("Server is listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}

	log.Println("Server stopped")
}
