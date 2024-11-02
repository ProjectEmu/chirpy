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
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
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
