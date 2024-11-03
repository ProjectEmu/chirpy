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
	case http.MethodDelete:
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (cfg *apiConfig) handlerChirpByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg.handleGetChirpByID(w, r)
	case http.MethodDelete:
		cfg.handleDeleteChirpByID(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusUnauthorized)
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

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
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

func (cfg *apiConfig) handleDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the chirp ID from the URL path
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

	// Extract Bearer Token from request header
	bearer, err := authy.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Issue parsing bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid bearer token")
		return
	}

	// Validate JWT and get User ID
	userID, err := authy.ValidateJWT(bearer, cfg.JWTSecret)
	if err != nil {
		log.Printf("Issue authenticating bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Retrieve chirp from the database by ID to ensure it exists and belongs to the user
	chirp, err := cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Chirp not found, return 404
			http.Error(w, "Chirp not found", http.StatusNotFound)
			return
		}
		// Some other error occurred, return 500
		log.Printf("Error retrieving chirp by ID: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Check if the authenticated user is the author of the chirp
	if chirp.UserID != userID {
		log.Printf("User %s attempted to delete chirp %s, which does not belong to them", userID, chirpID)
		respondWithError(w, http.StatusForbidden, "You are not allowed to delete this chirp")
		return
	}

	// Use SQLC's DeleteChirp method to delete the chirp
	err = cfg.DB.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error deleting chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not delete chirp")
		return
	}

	// Respond with 204 No Content if deletion was successful
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Log the raw query values for verification
	log.Printf("Received query parameters: %v", r.URL.Query())

	// Extract query parameters for author_id and sort
	authorIDParam := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")

	log.Printf("Received GET /api/chirps request with author_id: %s, sort: %s", authorIDParam, sortOrder)

	// Set default sorting order if not provided
	if sortOrder == "" {
		sortOrder = "asc"
	} else if sortOrder != "asc" && sortOrder != "desc" {
		log.Printf("Invalid sort parameter: %s", sortOrder)
		respondWithError(w, http.StatusBadRequest, "Invalid sort parameter, must be 'asc' or 'desc'")
		return
	}
	queryParams := database.GetChirpsWithFilterAndSortParams{}
	// Parse author ID if provided
	var authorID uuid.UUID
	if authorIDParam != "" {
		parsedUUID, err := uuid.Parse(authorIDParam)
		if err != nil {
			log.Printf("Invalid author_id provided: %s", err)
			respondWithError(w, http.StatusBadRequest, "Invalid author_id")
			return
		}
		log.Printf("Successfully parsed author_id: %s", parsedUUID)
		authorID = parsedUUID
		queryParams.Column1 = authorID
	} else {
		log.Println("No author_id provided, defaulting to empty UUID.")
		queryParams.Column1 = uuid.Nil
	}
	queryParams.Column2 = sortOrder
	log.Printf("Parsed author_id: %s, sort_order: %s", authorID, sortOrder)

	log.Printf("Query Params Struct: %+v", queryParams)

	// Execute the SQL query with appropriate parameters
	chirps, err := cfg.DB.GetChirpsWithFilterAndSort(r.Context(), queryParams)
	if err != nil {
		log.Printf("Error retrieving chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps")
		return
	}

	log.Printf("Number of chirps retrieved: %d", len(chirps))

	// Prepare the response
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

	log.Printf("Returning %d chirps in response", len(responseChirps))

	respondWithJSON(w, http.StatusOK, responseChirps)
}
