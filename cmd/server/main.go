package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"event-router-auth/internal/dispatcher"
	"event-router-auth/internal/handler"
)

func main() {
	logger := log.New(os.Stdout, "[event-router] ", log.LstdFlags|log.Lshortfile)

	// Initialize forwarder
	forwarder := dispatcher.NewForwarder(logger)

	mux := http.NewServeMux()

	// Register HTTP handlers
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/webhook", handler.WebhookHandler(forwarder))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Println("🚀 Starting event-router-auth on :8080")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("❌ Server failed: %v", err)
	}
}
