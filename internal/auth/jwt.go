package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	issuers   map[string]string
	hs256Keys map[string][]byte
	rs256Keys map[string]*rsa.PublicKey
	logger    *log.Logger
}

func NewJWTValidator(logger *log.Logger) *JWTValidator {
	return &JWTValidator{
		logger: logger,
		issuers: map[string]string{
			"https://oauth2.googleapis.com":               "event-router",
			"https://token.actions.githubusercontent.com": "event-router",
		},
		hs256Keys: map[string][]byte{
			"https://oauth2.googleapis.com": []byte("your-dev-shared-secret"),
		},
		rs256Keys: map[string]*rsa.PublicKey{
			"https://token.actions.githubusercontent.com": mustLoadRSAPublicKey("config/github-actions.pub"),
		},
	}
}

func mustLoadRSAPublicKey(path string) *rsa.PublicKey {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("Failed to read RSA public key: " + err.Error())
	}

	block, _ := pem.Decode(data)
	if block == nil {
		panic("Invalid PEM data")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("Invalid RSA public key: " + err.Error())
	}

	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		panic("Provided key is not RSA public key")
	}

	log.Printf("[auth] Loaded RSA public key from %s", path)
	return key
}

func (v *JWTValidator) Validate(r *http.Request) error {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("missing or invalid Authorization header")
	}

	rawToken := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid claims format")
		}

		iss, ok := claims["iss"].(string)
		if !ok {
			return nil, errors.New("missing issuer")
		}

		switch token.Method.Alg() {
		case jwt.SigningMethodHS256.Alg():
			key, ok := v.hs256Keys[iss]
			if !ok {
				return nil, fmt.Errorf("no HS256 key for issuer: %s", iss)
			}
			return key, nil

		case jwt.SigningMethodRS256.Alg():
			key, ok := v.rs256Keys[iss]
			if !ok {
				return nil, fmt.Errorf("no RS256 key for issuer: %s", iss)
			}
			return key, nil

		default:
			return nil, fmt.Errorf("unsupported signing method: %v", token.Method.Alg())
		}
	})

	if err != nil {
		return fmt.Errorf("JWT parse error: %w", err)
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)

	iss, _ := claims["iss"].(string)
	aud, _ := claims["aud"].(string)
	exp, _ := claims["exp"].(float64)
	sub, _ := claims["sub"].(string)

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

	v.logger.Printf("✅ JWT (alg: %s) validated for sub=%s, iss=%s", token.Method.Alg(), sub, iss)
	return nil
}

// ValidateJWT_RS256 returns the `sub` from RS256 token, useful when caller needs user identity
func ValidateJWT_RS256(tokenString string, expectedAud string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return mustLoadRSAPublicKey("config/github-actions.pub"), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		iss, _ := claims["iss"].(string)
		aud, _ := claims["aud"].(string)
		sub, _ := claims["sub"].(string)

		if !strings.HasPrefix(iss, "https://") {
			return "", errors.New("invalid issuer")
		}
		if aud != expectedAud {
			return "", errors.New("invalid audience")
		}
		return sub, nil
	}
	return "", errors.New("invalid JWT")
}
