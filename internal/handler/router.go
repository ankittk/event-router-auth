package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"event-router-auth/internal/dispatcher"
	"event-router-auth/internal/model"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ok"))
}

func WebhookHandler(forwarder dispatcher.Forwarder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var event model.Event
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "invalid event payload", http.StatusBadRequest)
			return
		}

		if err := forwarder.Forward(event); err != nil {
			log.Printf("error forwarding event: %v", err)
			http.Error(w, "failed to forward event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("event received"))
	}
}
