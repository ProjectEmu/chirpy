package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	authy "github.com/ProjectEmu/chirpy/internal/auth"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PolkaWebhookRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

// Main handler
func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	apiKey, err := authy.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("Error authorizing webhook request: %s", err)
		http.Error(w, "Unauthorized Request", http.StatusUnauthorized)
		return
	}
	if apiKey != cfg.Polka_apiKey {
		log.Printf("We chedk us up the key: %s    with stored: %s", apiKey, cfg.Polka_apiKey)
		http.Error(w, "Unauthorized Request", http.StatusUnauthorized)
		return
	}

	var req PolkaWebhookRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding webhook request: %s", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Event != "user.upgraded" {
		// For any event other than "user.upgraded", we respond with 204 No Content
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Parse the user ID from the request
	userID, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		log.Printf("Invalid user ID in webhook request: %s", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Upgrade the user to Chirpy Red using SQLC
	err = cfg.DB.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found, return 404
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		// Some other error occurred, return 500
		log.Printf("Error upgrading user to Chirpy Red: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}
