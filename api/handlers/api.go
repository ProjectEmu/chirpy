package handlers

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/ProjectEmu/chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	JWTSecret      string
}

type errorResponse struct {
	Error string `json:"error"`
}

type validResponse struct {
	Valid bool `json:"valid"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func SetupRoutes(mux *http.ServeMux, dbQueries *database.Queries, platform string, JWTSecret string) {

	apiCfg := &apiConfig{}
	apiCfg.DB = dbQueries
	apiCfg.Platform = platform
	apiCfg.JWTSecret = JWTSecret

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("/api/chirps", apiCfg.handlerChirps)
	mux.HandleFunc("/api/chirps/", apiCfg.handlerChirpByID)
	mux.HandleFunc("/api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("/api/login", apiCfg.handlerUserLogin)

}
