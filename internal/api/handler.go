package api

import (
	"encoding/json"
	"net/http"

	"github.com/jamesmeyerr/credit-card-validator/internal/luhn"
)

// Request represents the JSON request structure
type Request struct {
	CardNumber string `json:"card_number"`
}

// Response represents the JSON response structure
type Response struct {
	Valid bool `json:"valid"`
}

// ValidationHandler handles credit card validation requests
func ValidationHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse the request
	var req Request
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON payload"})
		return
	}

	// Validate the card number
	isValid := luhn.IsValid(req.CardNumber)

	// Prepare response
	resp := Response{
		Valid: isValid,
	}

	// Return response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}