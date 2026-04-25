package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NishDaswani/Chirpy/internal/auth"
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
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Defer closing the body & create request var
	defer r.Body.Close()
	var req Request
	// Decode into req variable
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	hashedPsw, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "issue hashing password")
		return
	}

	params := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPsw,
	}

	// Create new User in DB using req var, use r.Context() for cancel/timeout handling
	dbUser, err := cfg.db.CreateUser(r.Context(), params)
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

	// Authentication
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to get authentication token")
		return
	}

	gotUserID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "user is unauthorized")
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
		UserID: gotUserID,
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
		UserID:    gotUserID,
	}
	respondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	// Return error if incorrect Method
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting chirps")
		return
	}

	chirps := make([]Chirp, 0, len(dbChirps))

	for _, chirp := range dbChirps {
		temp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		chirps = append(chirps, temp)
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	// Return error if incorrect Method
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse chirp ID")
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "error getting chirp, chirp may not exist")
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) handlerValidateLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	type Request struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	var req Request

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	ok, err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
	if err != nil || !ok {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	tokenString, err := auth.MakeJWT(dbUser.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating access token")
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:  auth.MakeRefreshToken(),
		UserID: dbUser.ID,
	}

	dbRefreshToken, err := cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating refresh token")
		return
	}

	user := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        tokenString,
		RefreshToken: dbRefreshToken.Token,
	}

	respondWithJSON(w, http.StatusOK, user)

}

func (cfg *apiConfig) handlerValidateRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to get refresh token from authentication header")
		return
	}

	dbRefreshToken, err := cfg.db.GetUserByRefreshToken(r.Context(), refreshToken)
	if err != nil || dbRefreshToken.RevokedAt.Valid || time.Now().UTC().After(dbRefreshToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "error getting user information with refresh token")
		return
	}

	token, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating access token")
		return
	}
	type AccessToken struct {
		Token string `json:"token"`
	}

	accessToken := AccessToken{Token: token}

	respondWithJSON(w, http.StatusOK, accessToken)

}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to get refresh token from authentication header")
		return
	}
	if nrows, err := cfg.db.RevokeRefreshToken(r.Context(), refreshToken); err != nil || nrows == 0 {
		respondWithError(w, http.StatusBadRequest, "refresh token not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpdateLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to get access token from authentication header")
		return
	}

	gotUserID, err := auth.ValidateJWT(accessToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "user is unauthorized")
		return
	}

	type UpdateVars struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	var req UpdateVars

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	hashedPsw, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "issue hashing password")
		return
	}

	dbUpdateVars := database.UpdateLoginByUserIDParams{
		ID:             gotUserID,
		Email:          req.Email,
		HashedPassword: hashedPsw,
	}

	dbUser, err := cfg.db.UpdateLoginByUserID(r.Context(), dbUpdateVars)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error updating user")
		return
	}

	type UpdatedUserResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	user := UpdatedUserResponse{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	// Authentication
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to get authentication token")
		return
	}

	gotUserID, err := auth.ValidateJWT(accessToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "user is unauthorized")
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse chirp ID")
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "error getting chirp, chirp may not exist")
		return
	}

	if dbChirp.UserID != gotUserID {
		respondWithError(w, http.StatusForbidden, "user is not authorized to delete this chirp")
		return
	}

	if nrows, err := cfg.db.DeleteChirpByID(r.Context(), chirpID); err != nil || nrows == 0 {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
