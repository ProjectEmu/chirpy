package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ProjectEmu/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	JWTSecret      string
	Polka_apiKey   string
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
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	apiCfg := &apiConfig{}
	apiCfg.DB = dbQueries
	apiCfg.Platform = platform
	apiCfg.JWTSecret = JWTSecret
	apiCfg.Polka_apiKey = os.Getenv("POLKA_KEY")

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("/api/chirps", apiCfg.handlerChirps)
	mux.HandleFunc("/api/chirps/", apiCfg.handlerChirpByID)
	mux.HandleFunc("/api/users", apiCfg.handlerUsers)
	mux.HandleFunc("/api/login", apiCfg.handlerUserLogin)
	mux.HandleFunc("/api/revoke", apiCfg.handlerRevokeToken)
	mux.HandleFunc("/api/refresh", apiCfg.handlerRefreshToken)
	mux.HandleFunc("/api/polka/webhooks", apiCfg.handlerPolkaWebhook)
}
