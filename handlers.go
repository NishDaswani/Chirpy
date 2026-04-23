package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NishDaswani/Chirpy/internal/database"
	"github.com/google/uuid"
)

const adminMetrics = `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, adminMetrics, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) deleteUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if cfg.PLATFORM != "dev" {
		respondWithError(w, http.StatusForbidden, "not allowed to delete all users in non dev environment")
		return
	}
	err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error deleting users")
		return
	}
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// Return error if incorrect Method
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	// Create Request Struct
	type Request struct {
		Email string `json:"email"`
	}
	// Defer closing the body & create request var
	defer r.Body.Close()
	var req Request
	// Decode into req variable
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	// Create new User in DB using req var, use r.Context() for cancel/timeout handling
	dbUser, err := cfg.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating new user")
		return
	}

	usr := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, usr)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	// Return error if incorrect Method
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	defer r.Body.Close()

	type CreateChirpRequest struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	var post CreateChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	cleanPost, err := validateAndCleanChirp(post.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "chirp is too long")
		return
	}

	params := database.CreateChirpParams{
		Body:   cleanPost,
		UserID: post.UserID,
	}

	dbChirp, err := cfg.db.CreateChirp(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating chirp")
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
	respondWithJSON(w, http.StatusCreated, chirp)
}
