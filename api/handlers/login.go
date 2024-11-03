package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ProjectEmu/chirpy/config"
	authy "github.com/ProjectEmu/chirpy/internal/auth"
	"github.com/ProjectEmu/chirpy/internal/database"

	_ "github.com/lib/pq"
)

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Expires  int    `json:"expires"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	expires := config.AccessTokenDuration
	// Parse and check Expires
	if req.Expires > 0 && req.Expires < config.AccessTokenDuration {
		expires = req.Expires
	}

	// Use SQLC's AuthUser method
	user, err := cfg.DB.AuthUser(r.Context(), req.Email)
	if err == sql.ErrNoRows {
		log.Printf("User not found for email: %s", req.Email)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	} else if err != nil {
		log.Printf("Error retrieving user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve user")
		return
	}

	//Check Password
	err = authy.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		} else {
			log.Printf("Unexpected error during password hash check: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Could not verify password")
		}
		return
	}

	// Get access token
	token, err := authy.MakeJWT(user.ID, cfg.JWTSecret, time.Duration(expires)*time.Second)
	if err != nil {
		log.Printf("Could not fetch JWT: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not fetch JWT")
	}
	// Get refresh token
	refresh_token, err := authy.MakeRefreshToken()
	if err != nil {
		log.Printf("Could not fetch refresh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not fetch refresh token")
	}
	// Store refresh token
	expiryDate := time.Now().AddDate(0, 0, config.RefreshTokenDuration)
	request_tokenParams := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    user.ID,
		ExpiresAt: expiryDate,
	}
	_, err = cfg.DB.CreateRefreshToken(r.Context(), request_tokenParams)
	if err != nil {
		log.Printf("Could not store refresh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not store refresh token")
	}
	// Map database.User to the User struct to control JSON keys
	var responseUser struct {
		User
		Token         string `json:"token"`
		Refresh_Token string `json:"refresh_token"`
	}
	responseUser.ID = user.ID
	responseUser.CreatedAt = user.CreatedAt
	responseUser.UpdatedAt = user.UpdatedAt
	responseUser.Email = user.Email
	responseUser.Token = token
	responseUser.Refresh_Token = refresh_token
	responseUser.IsChirpyRed = user.IsChirpyRed

	respondWithJSON(w, http.StatusOK, responseUser)
}
