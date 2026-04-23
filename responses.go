package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	msgData, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(w, "Error with payload: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(msgData)
}

func respondWithCleanText(msg string) string {
	splitMsg := strings.Split(msg, " ")
	set := make(map[string]struct{})
	set["kerfuffle"] = struct{}{}
	set["sharbert"] = struct{}{}
	set["fornax"] = struct{}{}
	for idx, word := range splitMsg {
		if _, exists := set[strings.ToLower(word)]; exists {
			splitMsg[idx] = "****"
		}
	}
	cleanText := strings.Join(splitMsg, " ")
	return cleanText
}

func validateAndCleanChirp(post string) (string, error) {
	const errorMsg = "Chirp is too long"
	if len(post) > 140 {
		return "", fmt.Errorf(errorMsg)
	} else {
		return respondWithCleanText(post), nil
	}
}
