package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateChirp_CleansProfaneWords(t *testing.T) {
	// Build fake HTTP request
	body := `{"body": "This is a kerfuffle opinion I need to share with the world"}`
	req := httptest.NewRequest(http.MethodPost, "/api/validate_chirp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create fake response writer
	rr := httptest.NewRecorder()

	// Call handler
	handlerValidateChirp(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Assert JSON body

	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}
	var got response

	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	want := "This is a **** opinion I need to share with the world"

	if got.CleanedBody != want {
		t.Fatalf("expected cleaned_body %s, got %q", want, got.CleanedBody)
	}
}

func TestHandlerValidateChirp_TooLong(t *testing.T) {
	longText := strings.Repeat("a", 141)
	body := `{"body":"` + longText + `"}`

	req := httptest.NewRequest(http.MethodPost, "/api/validate_chirp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handlerValidateChirp(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	type errorResponse struct {
		Error string `json:"error"`
	}
	var got errorResponse
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if got.Error != "Chirp is too long" {
		t.Fatalf("expected error %q, got %q", "Chirp is too long", got.Error)
	}
}
