package main

import (
	"strings"
	"testing"
)

func TestValidateAndCleanChirp_CleansProfaneWords(t *testing.T) {
	input := "This is a kerfuffle opinion I need to share with the world"

	got, err := validateAndCleanChirp(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := "This is a **** opinion I need to share with the world"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestValidateAndCleanChirp_TooLong(t *testing.T) {
	input := strings.Repeat("a", 141)

	got, err := validateAndCleanChirp(input)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if err.Error() != "Chirp is too long" {
		t.Fatalf("expected error %q, got %q", "Chirp is too long", err.Error())
	}
	if got != "" {
		t.Fatalf("expected empty cleaned text, got %q", got)
	}
}
