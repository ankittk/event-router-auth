package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	issuer := "https://token.actions.githubusercontent.com"
	secret := []byte("your-gh-dev-secret")

	claims := jwt.MapClaims{
		"iss": issuer,
		"aud": "event-router",
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"sub": "github-actions[bot]",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		panic(err)
	}

	fmt.Print(signedToken)
}
