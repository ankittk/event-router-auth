package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
)

type HMACValidator struct {
	Secrets []string
}

func NewHMACValidator(secrets []string) *HMACValidator {
	return &HMACValidator{Secrets: secrets}
}

// Validate only checks HMAC signature, ignores timestamp
func (v *HMACValidator) Validate(r *http.Request, body []byte) error {
	sig := r.Header.Get("X-Signature-256")
	ghSig := r.Header.Get("X-Hub-Signature-256")

	if sig == "" && ghSig == "" {
		return errors.New("missing HMAC signature")
	}

	for _, secret := range v.Secrets {
		expected := computeHMACSHA256(body, secret)

		if sig != "" && hmac.Equal([]byte(expected), []byte(sig)) {
			return nil
		}

		ghExpected := "sha256=" + expected
		if ghSig != "" && hmac.Equal([]byte(ghExpected), []byte(ghSig)) {
			return nil
		}
	}

	return errors.New("invalid HMAC signature")
}

func computeHMACSHA256(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
