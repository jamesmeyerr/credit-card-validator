package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// SanitizationConfig defines sanitization rules
type SanitizationConfig struct {
	MaxCardNumberLength int
	MaxExpiryLength     int
	MaxCVVLength        int
	MaxRequestSize      int64 // in bytes
}

// DefaultSanitizationConfig returns a default configuration
func DefaultSanitizationConfig() SanitizationConfig {
	return SanitizationConfig{
		MaxCardNumberLength: 19,    // Maximum valid card number length
		MaxExpiryLength:     5,     // Format: MM/YY
		MaxCVVLength:        4,     // Max 4 digits for Amex
		MaxRequestSize:      1024,  // 1KB is more than enough for our small JSON payload
	}
}

// InputSanitizer middleware handles input sanitization
type InputSanitizer struct {
	config SanitizationConfig
}

// NewInputSanitizer creates a new input sanitization middleware
func NewInputSanitizer(config SanitizationConfig) *InputSanitizer {
	return &InputSanitizer{
		config: config,
	}
}

// SanitizeMiddleware creates a middleware function for input sanitization
func (is *InputSanitizer) SanitizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only process POST/GET requests with JSON content
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(strings.ToLower(contentType), "application/json") {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		// Limit request size
		r.Body = http.MaxBytesReader(w, r.Body, is.config.MaxRequestSize)
		
		// Read the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		
		// Close the original body
		r.Body.Close()
		
		// Try to parse as JSON to ensure it's valid
		var requestMap map[string]interface{}
		if err := json.Unmarshal(body, &requestMap); err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Sanitize card number - only keep digits
		if cardNumber, ok := requestMap["card_number"].(string); ok {
			sanitized := sanitizeCardNumber(cardNumber)
			if len(sanitized) > is.config.MaxCardNumberLength {
				http.Error(w, "Card number exceeds maximum allowed length", http.StatusBadRequest)
				return
			}
			requestMap["card_number"] = sanitized
		}

		// Sanitize expiry date - validate format
		if expiryDate, ok := requestMap["expiry_date"].(string); ok {
			if !isValidExpiryFormat(expiryDate) || len(expiryDate) > is.config.MaxExpiryLength {
				http.Error(w, "Invalid expiry date format", http.StatusBadRequest)
				return
			}
		}

		// Sanitize CVV - only allow digits
		if cvv, ok := requestMap["cvv"].(string); ok {
			if !isValidCVV(cvv) || len(cvv) > is.config.MaxCVVLength {
				http.Error(w, "Invalid CVV format", http.StatusBadRequest)
				return
			}
		}

		// Convert back to JSON
		sanitizedBody, err := json.Marshal(requestMap)
		if err != nil {
			http.Error(w, "Error processing request", http.StatusInternalServerError)
			return
		}

		// Replace the request body with our sanitized version
		r.Body = io.NopCloser(bytes.NewBuffer(sanitizedBody))
		
		// Update Content-Length header
		r.ContentLength = int64(len(sanitizedBody))
		
		// Pass to next handler
		next.ServeHTTP(w, r)
	})
}

// sanitizeCardNumber removes all non-digit characters
func sanitizeCardNumber(input string) string {
	var sanitized strings.Builder
	for _, char := range input {
		if char >= '0' && char <= '9' {
			sanitized.WriteRune(char)
		}
	}
	return sanitized.String()
}

// isValidExpiryFormat checks if expiry date follows MM/YY format
func isValidExpiryFormat(input string) bool {
	pattern := regexp.MustCompile(`^(0[1-9]|1[0-2])/\d{2}$`)
	return pattern.MatchString(input)
}

// isValidCVV checks if CVV contains only digits
func isValidCVV(input string) bool {
	pattern := regexp.MustCompile(`^\d{3,4}$`)
	return pattern.MatchString(input)
}