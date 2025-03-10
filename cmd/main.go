package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goel-aayush/students-api/internal/config"
	"github.com/goel-aayush/students-api/internal/http/handlers/student"
	"github.com/goel-aayush/students-api/internal/storage/sqlite"
	"github.com/gorilla/mux" // Use gorilla/mux for routing
)

func main() {
	// Load config
	cfg := config.MustLoad()

	// Database setup
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage initialized", slog.String("env", cfg.Env))

	// Setup router using gorilla/mux
	router := mux.NewRouter()

	router.HandleFunc("/api/students", student.New(storage)).Methods("POST")
	router.HandleFunc("/api/students/{id}", student.GetById(storage)).Methods("GET")
	router.HandleFunc("/api/students", student.GetList(storage)).Methods("GET")
	router.HandleFunc("/api/students/{id}", student.UpdateStudent(storage)).Methods("PATCH")
	router.HandleFunc("/api/students/{id}", student.RemoveStudent(storage)).Methods("DELETE")

	// Setup server
	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	slog.Info("server started", slog.String("address", cfg.Addr))

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server: ", err)
		}
	}()

	<-done

	slog.Info("shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server shutdown successfully")
}
