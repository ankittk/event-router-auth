package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"event-router-auth/internal/dispatcher"
	"event-router-auth/internal/model"
)

const (
	signatureHeader       = "X-Signature-256"
	githubSignatureHeader = "X-Hub-Signature-256"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func WebhookHandler(forwarder dispatcher.Forwarder) http.HandlerFunc {
	secret := os.Getenv("HMAC_SECRET")
	if secret == "" {
		log.Fatalln("HMAC_SECRET not set")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sig := r.Header.Get(signatureHeader)
		ghSig := r.Header.Get(githubSignatureHeader)
		if sig == "" && ghSig == "" {
			http.Error(w, "missing signature header", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		var mac []byte
		secretKeys := strings.Split(secret, ",")
		valid := false
		for _, s := range secretKeys {
			m := hmac.New(sha256.New, []byte(strings.TrimSpace(s)))
			m.Write(body)
			mac = m.Sum(nil)
			expectedSig := hex.EncodeToString(mac)
			if sig != "" && hmac.Equal([]byte(expectedSig), []byte(sig)) {
				valid = true
				break
			}
			ghExpected := "sha256=" + expectedSig
			if ghSig != "" && hmac.Equal([]byte(ghExpected), []byte(ghSig)) {
				valid = true
				break
			}
		}

		if !valid {
			http.Error(w, "invalid HMAC signature", http.StatusUnauthorized)
			return
		}

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

func JWTWebhookHandler(forwarder dispatcher.Forwarder) http.HandlerFunc {
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
		_ = r.Body.Close()

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

func CLISendEvent(forwarder dispatcher.Forwarder) http.HandlerFunc {
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

		user := GetUserFromContext(r.Context())

		log.Printf("Received CLI event from user: %s | Event: %+v", user, event)

		if err := forwarder.Forward(event); err != nil {
			log.Printf("Error forwarding event from user %s: %v", user, err)
			http.Error(w, "failed to forward event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("event accepted from CLI"))
	}
}
