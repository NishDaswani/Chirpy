package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	type parameters struct {
		Body string `json:"body"`
	}
	params := parameters{}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	type returnVals struct {
		CleanedText string `json:"cleaned_body"`
	}
	const errorMsg = "Chirp is too long"
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, errorMsg)
	} else {
		respondWithJSON(w, http.StatusOK, returnVals{CleanedText: respondWithCleanText(params.Body)})
	}
}
