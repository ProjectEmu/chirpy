package handlers

import (
	"database/sql"
	"log"
	"net/http"

	authy "github.com/ProjectEmu/chirpy/internal/auth"

	_ "github.com/lib/pq"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Bearer Token
	refreshToken, err := authy.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Issue parsing bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization token")
		return
	}

	// Revoke Refresh Token
	err = cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
	if err == sql.ErrNoRows {
		log.Printf("Attempted to revoke non-existent refresh token: %s", refreshToken)
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization token")
		return
	} else if err != nil {
		log.Printf("Database error while revoking refresh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Unable to revoke token at this time")
		return
	}

	// Log success
	log.Printf("Successfully revoked refresh token: %s", refreshToken)

	// Respond with a status indicating success
	w.WriteHeader(http.StatusNoContent)
}
