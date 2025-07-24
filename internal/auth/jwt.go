package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	issuers map[string]string
	keys    map[string][]byte
	logger  *log.Logger
}

func NewJWTValidator(logger *log.Logger) *JWTValidator {
	return &JWTValidator{
		logger: logger,
		issuers: map[string]string{
			"https://oauth2.googleapis.com":               "event-router",
			"https://token.actions.githubusercontent.com": "event-router",
		},
		keys: map[string][]byte{
			"https://oauth2.googleapis.com":               []byte("your-dev-shared-secret"),
			"https://token.actions.githubusercontent.com": []byte("your-gh-dev-secret"),
		},
	}
}

func (v *JWTValidator) Validate(r *http.Request) error {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("missing or invalid Authorization header")
	}

	rawToken := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid claims format")
		}
		iss, ok := claims["iss"].(string)
		if !ok {
			return nil, errors.New("missing issuer claim")
		}
		key, ok := v.keys[iss]
		if !ok {
			return nil, fmt.Errorf("unknown issuer: %s", iss)
		}
		return key, nil
	})
	if err != nil {
		return fmt.Errorf("JWT parse error: %w", err)
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)

	iss, ok := claims["iss"].(string)
	if !ok {
		return errors.New("missing issuer claim")
	}

	aud, ok := claims["aud"].(string)
	if !ok {
		return errors.New("missing audience claim")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("missing or invalid expiration")
	}

	expectedAud, exists := v.issuers[iss]
	if !exists {
		return fmt.Errorf("untrusted issuer: %s", iss)
	}

	if aud != expectedAud {
		return fmt.Errorf("invalid audience: got %s, expected %s", aud, expectedAud)
	}

	if time.Now().Unix() > int64(exp) {
		return errors.New("token expired")
	}

	v.logger.Printf("✅ JWT validated for issuer %s", iss)
	return nil
}
