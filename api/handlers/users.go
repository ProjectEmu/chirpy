package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	authy "github.com/ProjectEmu/chirpy/internal/auth"
	"github.com/ProjectEmu/chirpy/internal/database"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

// Main handler
func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		cfg.handleCreateUser(w, r)
	case http.MethodPut:
		cfg.handleUpdateUser(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	//Prepare parameters
	pwHash, err := authy.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error creating password hash: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid password")
		return
	}
	userParams := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: pwHash,
	}

	// Use SQLC's CreateUser method
	user, err := cfg.DB.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	// Map database.User to the User struct to control JSON keys
	responseUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, responseUser)
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Bearer Token
	bearer, err := authy.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Issue parsing bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid bearer token")
		return
	}

	// Validate JWT and get User ID
	userID, err := authy.ValidateJWT(bearer, cfg.JWTSecret)
	if err != nil {
		log.Printf("Issue authenticating bearer: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired access token")
		return
	}

	// Decode request body to get new email and password
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Hash the new password
	pwHash, err := authy.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error creating password hash: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid password")
		return
	}

	// Prepare parameters for update
	updateParams := database.UpdateUserParams{
		ID:             userID,
		Email:          req.Email,
		HashedPassword: pwHash,
	}

	// Update user using SQLC's UpdateUser method
	updatedUser, err := cfg.DB.UpdateUser(r.Context(), updateParams)
	if err != nil {
		log.Printf("Error updating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not update user")
		return
	}

	// Map updated user to response format
	responseUser := User{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email:     updatedUser.Email,
	}

	respondWithJSON(w, http.StatusOK, responseUser)
}
