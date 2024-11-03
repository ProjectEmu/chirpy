package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/ProjectEmu/chirpy/config"
	authy "github.com/ProjectEmu/chirpy/internal/auth"

	_ "github.com/lib/pq"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Bearer Token
	refreshToken, err := authy.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Issue parsing bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid bearer token")
		return
	}

	// Validate Refresh Token
	refreshTokenResult, err := cfg.DB.GetRefreshToken(r.Context(), refreshToken)
	if err == sql.ErrNoRows {
		log.Printf("Refresh token not found: %s", refreshToken)
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	} else if err != nil {
		log.Printf("Error retrieving refresh token from database: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Unable to validate refresh token")
		return
	}

	// Check if the token is expired or revoked
	if refreshTokenResult.ExpiresAt.Before(time.Now()) {
		log.Println("Refresh token is expired.")
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	if refreshTokenResult.RevokedAt.Valid {
		log.Println("Refresh token has been revoked.")
		respondWithError(w, http.StatusUnauthorized, "Refresh token revoked")
		return
	}

	// Generate a new access token
	accessToken, err := authy.MakeJWT(refreshTokenResult.UserID, cfg.JWTSecret, time.Duration(config.AccessTokenDuration)*time.Second)
	if err != nil {
		log.Printf("Could not generate access token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not generate access token")
		return
	}

	// Optionally revoke the used refresh token
	//revokeErr := cfg.DB.RevokeRefreshToken(r.Context(), refreshTokenResult.Token, time.Now())
	//if revokeErr != nil {
	//    log.Printf("Error revoking refresh token: %s", revokeErr)
	//    // Not returning here; the error is logged but does not prevent a successful access token response
	//}

	var responseToken struct {
		Token string `json:"token"`
	}
	responseToken.Token = accessToken

	respondWithJSON(w, http.StatusOK, responseToken)
}
