package api

import (
	"encoding/json"
	"net/http"

	"github.com/jamesmeyerr/credit-card-validator/internal/luhn"
)

// Request represents the JSON request structure
type Request struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date,omitempty"` // Format: MM/YY
}

// Response represents the JSON response structure
type Response struct {
	Valid           bool   `json:"valid"`
	Network         string `json:"network,omitempty"`
	CardLength      int    `json:"card_length,omitempty"`
	ExpiryValid     bool   `json:"expiry_valid,omitempty"`
	ExpiryFormatOK  bool   `json:"expiry_format_ok,omitempty"`
	Message         string `json:"message,omitempty"`
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
	if req.CardNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Card number is required"})
		return
	}

	// Create validation request
	validationReq := luhn.CardValidationRequest{
		CardNumber: req.CardNumber,
		ExpiryDate: req.ExpiryDate,
	}

	// Get card information
	cardInfo := luhn.ValidateCard(validationReq)

	// Prepare response message
	message := buildResponseMessage(cardInfo)

	// Prepare response
	resp := Response{
		Valid:          cardInfo.Valid,
		Network:        cardInfo.Network,
		CardLength:     cardInfo.CardLength,
		ExpiryValid:    cardInfo.ExpiryValid,
		ExpiryFormatOK: cardInfo.ExpiryFormatOK,
		Message:        message,
	}

	// Return response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// buildResponseMessage creates a human-readable message based on validation results
func buildResponseMessage(cardInfo luhn.CardInfo) string {
	if !cardInfo.Valid {
		return "Card number is invalid (failed Luhn check)"
	}

	networkInfo := "unknown network"
	if cardInfo.Network != "Unknown" {
		networkInfo = cardInfo.Network
	}

	message := "Valid " + networkInfo + " card"

	// Add expiry information if provided
	if cardInfo.ExpiryFormatOK {
		if cardInfo.ExpiryValid {
			message += " with valid expiration date"
		} else {
			message += " with expired or invalid expiration date"
		}
	}

	return message
}