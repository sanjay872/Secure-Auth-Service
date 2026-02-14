package main

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTCreationAndValidation(t *testing.T) {

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(1 * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Parse token
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		t.Fatal("Token should be valid")
	}

	claims := parsedToken.Claims.(jwt.MapClaims)

	if claims["sub"] != "test-user" {
		t.Fatal("User ID mismatch")
	}
}

func TestExpiredJWT(t *testing.T) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "expired-user",
		"exp": time.Now().Add(-1 * time.Minute).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err == nil && parsedToken.Valid {
		t.Fatal("Expired token should not be valid")
	}
}
