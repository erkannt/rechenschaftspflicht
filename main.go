package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erkannt/rechenschaftspflicht/services/authentication"
	"github.com/erkannt/rechenschaftspflicht/services/config"
	database "github.com/erkannt/rechenschaftspflicht/services/db"
	"github.com/erkannt/rechenschaftspflicht/services/eventstore"
	"github.com/erkannt/rechenschaftspflicht/services/userstore"
	"github.com/julienschmidt/httprouter"
)

func run(
	ctx context.Context,
	getenv func(string) string,
	stdout io.Writer,
) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Setup dependencies
	db, err := database.InitDB()
	if err != nil {
		return fmt.Errorf("could not init database: %w", err)
	}
	cfg, err := config.LoadFromEnv(getenv)
	eventStore := eventstore.NewEventStore(db)
	userStore := userstore.NewUserStore(db)
	auth := authentication.New(cfg)

	// Create server
	router := httprouter.New()
	addRoutes(router, eventStore, userStore, auth)
	srv := &http.Server{Addr: ":8080", Handler: router}

	// Start the server
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		} else {
			serverErr <- nil
		}
	}()
	fmt.Fprintln(stdout, "Server is listening on :8080")

	// Graceful shutdown
	select {
	case <-ctx.Done():
		fmt.Fprintln(stdout, "Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}
		fmt.Fprintln(stdout, "Server stopped")
		return nil
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}
		fmt.Fprintln(stdout, "Server stopped")
		return nil
	}
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
