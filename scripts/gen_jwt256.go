package main

import (
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Load private key from the config directory (make sure it's the correct PEM format)
	privateKeyData, err := os.ReadFile("config/private_pkcs8.pem") // Update to your correct file path
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		panic("failed to parse private key PEM")
	}

	// If the key is not valid, it'll fail here
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(block.Bytes)
	if err != nil {
		panic(fmt.Sprintf("Error parsing private key: %v", err))
	}

	// Create the JWT claims
	claims := jwt.MapClaims{
		"iss": "https://oauth2.googleapis.com",                      // Issuer
		"aud": "event-router-auth",                                  // Audience
		"sub": "cli-user",                                           // Subject (user)
		"exp": jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // Expiration
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign the token with the private key
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		panic(fmt.Sprintf("Error signing JWT: %v", err))
	}

	// Output the JWT
	fmt.Println("Generated JWT:", signedToken)
}
