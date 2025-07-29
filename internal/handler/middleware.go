package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"event-router-auth/internal/auth"
)

type ctxKey string

const UserCtxKey ctxKey = "user"

func GetUserFromContext(ctx context.Context) string {
	if u := ctx.Value(UserCtxKey); u != nil {
		if str, ok := u.(string); ok {
			return str
		}
	}
	return "unknown"
}

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

func JWTMiddleware(validator *auth.JWTValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "unauthorized: missing bearer token", http.StatusUnauthorized)
			return
		}

		rawToken := strings.TrimPrefix(authHeader, "Bearer ")

		if sub, err := auth.ValidateJWT_RS256(rawToken, "event-router"); err == nil {
			ctx := context.WithValue(r.Context(), UserCtxKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if err := validator.Validate(r); err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequirePATAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.ValidatePAT(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
