package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	authy "github.com/ProjectEmu/chirpy/internal/auth"
	"github.com/ProjectEmu/chirpy/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User_id   uuid.UUID `json:"user_id"`
}

type chirpRequest struct {
	Body string `json:"body"`
}

// Main handler
func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		cfg.handleCreateChirp(w, r)
	case http.MethodGet:
		cfg.handleGetAllChirps(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func cleanProfanity(body string) string {
	profanities := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(body, " ")
	for i, word := range words {
		for _, profanity := range profanities {
			if strings.EqualFold(word, profanity) {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	bearer, err := authy.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Issue parsing bearer: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Issue parsing bearer")
		return
	}

	var req chirpRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(req.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	userID, err := authy.ValidateJWT(bearer, cfg.JWTSecret)
	if err != nil {
		log.Printf("Issue authenticating bearer: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Issue authenticating bearer")
		return
	}

	cleanedBody := cleanProfanity(req.Body)

	// Use SQLC's CreateChirp method

	chirpParams := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	}
	chirp, err := cfg.DB.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not chirp")
		return
	}

	// Map database.User to the User struct to control JSON keys
	responseChirp := Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		User_id:   chirp.UserID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	}

	respondWithJSON(w, http.StatusCreated, responseChirp)
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	chirps, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error retrieving all chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve all chirps")
		return
	}

	responseChirps := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		responseChirps[i] = Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			User_id:   chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}
	}

	respondWithJSON(w, http.StatusOK, responseChirps)
}

func (cfg *apiConfig) handlerChirpByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the ID from the URL path
	id := strings.TrimPrefix(r.URL.Path, "/api/chirps/")
	if id == "" {
		http.Error(w, "Missing chirp ID", http.StatusBadRequest)
		return
	}

	// Parse the ID to UUID
	chirpID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	// Use SQLC to retrieve the chirp by ID
	chirp, err := cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Chirp not found, return 404
			http.Error(w, "Chirp not found", http.StatusNotFound)
			return
		}
		// Some other error occurred, return 500
		log.Printf("Error retrieving chirp by ID: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Map the database chirp to a response chirp
	responseChirp := Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		User_id:   chirp.UserID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	}

	// Send back the chirp in JSON format
	respondWithJSON(w, http.StatusOK, responseChirp)
}
