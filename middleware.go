package main

import (
	"log"
	"net/http"
	"sync/atomic"

	"github.com/NishDaswani/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	PLATFORM       string
	JWTSecret      string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		log.Printf("%s %s (hit #%d)", r.Method, r.URL.Path, cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}
