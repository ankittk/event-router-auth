package handler

import (
	"bytes"
	"io"
	"net/http"

	"event-router-auth/internal/auth"
)

func HMACMiddleware(validator *auth.HMACValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "unable to read request body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		if err := validator.Validate(r, bodyBytes); err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
