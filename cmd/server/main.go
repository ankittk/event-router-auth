package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"event-router-auth/internal/auth"
	"event-router-auth/internal/dispatcher"
	"event-router-auth/internal/handler"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	logger := log.New(os.Stdout, "[event-router] ", log.LstdFlags|log.Lshortfile)

	hmacSecretEnv := os.Getenv("HMAC_SECRET")
	if hmacSecretEnv == "" {
		logger.Fatal("❌ HMAC_SECRET not set")
	}

	hmacSecrets := strings.Split(hmacSecretEnv, ",")
	for i := range hmacSecrets {
		hmacSecrets[i] = strings.TrimSpace(hmacSecrets[i])
	}
	logger.Printf("Loaded %d HMAC secret(s)", len(hmacSecrets))

	if *debug {
		logger.Println("Debug mode enabled")
	}

	hmacValidator := auth.NewHMACValidator(hmacSecrets)
	jwtValidator := auth.NewJWTValidator(logger)

	forwarder := dispatcher.NewForwarder(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthCheck)

	mux.Handle("/webhook", handler.HMACMiddleware(
		hmacValidator,
		handler.WebhookHandler(forwarder),
	))

	mux.Handle("/secure-webhook", handler.JWTMiddleware(
		jwtValidator,
		handler.JWTWebhookHandler(forwarder),
	))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Println("Starting event-router-auth on :8080")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("❌ Server failed: %v", err)
	}
}
