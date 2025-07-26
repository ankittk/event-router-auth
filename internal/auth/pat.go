package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"
)

var validPATs map[string]string

func init() {
	validPATs = make(map[string]string)

	pats := os.Getenv("VALID_PAT")
	if pats == "" {
		validPATs["my-secret-pat"] = "cli-user"
		return
	}

	for _, entry := range strings.Split(pats, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
		token := parts[0]

		if token == "" {
			continue
		}

		username := "cli-user"
		if len(parts) == 2 && parts[1] != "" {
			username = parts[1]
		}

		validPATs[token] = username
	}
}

func ValidatePAT(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no Authorization header found")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid Authorization scheme")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, ok := validPATs[token]
	if !ok {
		return "", errors.New("invalid or expired PAT")
	}

	return user, nil
}
