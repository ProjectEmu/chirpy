package handlers

import (
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if cfg.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Forbidden in non-development environments")
		return
	}

	// Delete all chirps
	err := cfg.DB.DeleteAllChirps(r.Context())
	if err != nil {
		log.Printf("Error deleting chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not reset chirps")
		return
	}

	// Delete all refresh tokens
	err = cfg.DB.DeleteAllRefreshTokens(r.Context())
	if err != nil {
		log.Printf("Error deleting refresh tokens: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not reset refresh tokens")
		return
	}

	err = cfg.DB.DeleteAllUsers(r.Context())
	if err != nil {
		log.Printf("Error deleting users: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not reset users")
		return
	}

	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("All chirps, refresh tokens, and users deleted. Hits counter reset to 0"))
}
