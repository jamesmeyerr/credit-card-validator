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
	CVV        string `json:"cvv,omitempty"`         // 3 or 4 digits
}

// Response represents the JSON response structure
type Response struct {
	Valid           bool   `json:"valid"`
	Network         string `json:"network,omitempty"`
	CardLength      int    `json:"card_length,omitempty"`
	ExpiryValid     bool   `json:"expiry_valid,omitempty"`
	ExpiryFormatOK  bool   `json:"expiry_format_ok,omitempty"`
	CVVValid        bool   `json:"cvv_valid,omitempty"`
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
		CVV:        req.CVV,
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
		CVVValid:       cardInfo.CVVValid,
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

	// Add CVV information if validated
	if cardInfo.CVVValid {
		message += " and valid security code (CVV)"
	} else if cardInfo.CVVValid == false && cardInfo.Network != "" {
		// Only mention invalid CVV if one was provided (otherwise CVVValid would be false by default)
		if cardInfo.Network == "American Express" {
			message += " but invalid security code (should be 4 digits)"
		} else {
			message += " but invalid security code (should be 3 digits)"
		}
	}

	return message
}