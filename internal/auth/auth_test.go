package auth

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "super-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == password {
		t.Fatal("HashPassword returned raw password instead of hash")
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error: %v", err)
	}
	if !match {
		t.Fatal("expected password/hash to match")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	match, err := CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error for wrong password: %v", err)
	}
	if match {
		t.Fatal("expected wrong password not to match hash")
	}
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	_, err := CheckPasswordHash("any-password", "not-a-valid-argon2id-hash")
	if err == nil {
		t.Fatal("expected error for invalid hash format")
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "DSWNISH"
	duration := time.Minute

	tokenString, err := MakeJWT(userID, tokenSecret, duration)
	if err != nil {
		t.Fatalf("error creating tokenString with valid arguments: %v", err)
	}

	if tokenString == "" {
		t.Fatal("empty token string")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "DSWNISH"
	duration := time.Minute
	tokenString, err := MakeJWT(userID, tokenSecret, duration)
	if err != nil {
		t.Fatalf("error creating tokenString with valid arguments: %v", err)
	}

	gotUserID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("failed to validate userID from jwt: %v", err)
	}

	if gotUserID != userID {
		t.Fatalf("userID mismatch: got %v want %v", gotUserID, userID)
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenString, err := MakeJWT(userID, "correct-secret", time.Minute)
	if err != nil {
		t.Fatalf("failed to create jwt: %v", err)
	}

	_, err = ValidateJWT(tokenString, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong token secret")
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenString, err := MakeJWT(userID, "test-secret", -time.Second)
	if err != nil {
		t.Fatalf("failed to create expired jwt: %v", err)
	}

	_, err = ValidateJWT(tokenString, "test-secret")
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestGetBearerToken(t *testing.T) {
	wantToken := "IgQPn+1yc0FWT28U/pq3cKpC7x2H0jBi4v/Lntpnr8R8YU7gix3wyGds94oqbPgVPwcQ/YyADOfqF7XCoLJvIw=="
	header := make(http.Header)
	header.Set("Authorization", fmt.Sprintf("Bearer %s", wantToken))
	gotToken, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("problem getting bearer token: %v", err)
	}

	if gotToken == "" {
		t.Fatal("token empty")
	}

	if gotToken != wantToken {
		t.Fatalf("token mismatch, got: token %q, want: token %q", gotToken, wantToken)
	}
}

func TestGetBearerToken_MissingHeader(t *testing.T) {
	header := make(http.Header)
	_, err := GetBearerToken(header)
	if err == nil {
		t.Fatal("missing header, expected error")
	}
	if !strings.Contains(err.Error(), "missing authorization header") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestGetBearerToken_MissingBearerPrefix(t *testing.T) {
	wantToken := "IgQPn+1yc0FWT28U/pq3cKpC7x2H0jBi4v/Lntpnr8R8YU7gix3wyGds94oqbPgVPwcQ/YyADOfqF7XCoLJvIw=="
	header := make(http.Header)
	header.Set("Authorization", wantToken)
	_, err := GetBearerToken(header)
	if err == nil {
		t.Fatal("expected error, missing bearer prefix")
	}
	if !strings.Contains(err.Error(), "does not contain 'Bearer'") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestGetBearerToken_EmptyBearerToken(t *testing.T) {
	header := make(http.Header)
	header.Set("Authorization", "Bearer    ")

	_, err := GetBearerToken(header)
	if err == nil {
		t.Fatal("expected error for empty bearer token")
	}
	if !strings.Contains(err.Error(), "empty Bearer token") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
