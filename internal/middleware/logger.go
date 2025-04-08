package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Initialize global logger
func init() {
	// Pretty console output for development
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

	// Set global log level (can be overridden by environment)
	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(logLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		// Default to info level
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// contextKey is a type for context keys used by the logger
type contextKey int

const (
	// requestIDKey is the context key for the request ID
	requestIDKey contextKey = iota
)

// LoggingMiddleware adds request logging and tracing
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate request ID if it doesn't exist
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		// Store request ID in context
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Create a response recorder to capture response data
		rr := &responseRecorder{
			ResponseWriter: w,
			Status:         http.StatusOK, // Default status
			Size:           0,
		}

		// Extract request body for logging (with privacy protection)
		var requestBody map[string]interface{}
		var bodyBytes []byte
		
		if r.Body != nil && r.Header.Get("Content-Type") == "application/json" {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body.Close()
			
			// Restore the body for downstream handlers
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			
			// Try to parse as JSON
			if err := json.Unmarshal(bodyBytes, &requestBody); err == nil {
				// Mask sensitive data
				if cardNum, ok := requestBody["card_number"].(string); ok && len(cardNum) > 6 {
					masked := cardNum[:6] + strings.Repeat("*", len(cardNum)-10) + cardNum[len(cardNum)-4:]
					requestBody["card_number"] = masked
				}
				if cvv, ok := requestBody["cvv"].(string); ok {
					requestBody["cvv"] = strings.Repeat("*", len(cvv))
				}
			}
		}

		// Pre-request logging
		logger := log.With().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("client_ip", getClientIP(r)).
			Str("user_agent", r.UserAgent()).
			Interface("query", r.URL.Query()).
			Logger()

		if requestBody != nil {
			logger = logger.With().Interface("request_body", requestBody).Logger()
		}

		logger.Info().Msg("Request started")

		// Process the request
		next.ServeHTTP(rr, r)

		// Post-request logging
		duration := time.Since(start)
		responseLog := logger.With().
			Int("status", rr.Status).
			Int("size", rr.Size).
			Dur("duration_ms", duration).
			Logger()

		if rr.Status >= 400 {
			// Log elevated for errors
			responseLog.Error().Msg("Request failed")
		} else {
			responseLog.Info().Msg("Request completed")
		}
	})
}

// responseRecorder is a wrapper around http.ResponseWriter to capture status code and response size
type responseRecorder struct {
	http.ResponseWriter
	Status int
	Size   int
}

// WriteHeader captures the status code
func (r *responseRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write captures the response size
func (r *responseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.Size += size
	return size, err
}

// getClientIP extracts the client's IP address from the request
func getClientIP(r *http.Request) string {
	// Try different headers that might contain the real client IP
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		ip := r.Header.Get(header)
		if ip != "" {
			// X-Forwarded-For can contain multiple IPs; use the first one
			return strings.Split(ip, ",")[0]
		}
	}
	
	// Fall back to remote address
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

// GetRequestID extracts the request ID from the context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// ApplicationLogger returns a logger instance with request context
func ApplicationLogger(ctx context.Context) zerolog.Logger {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		return log.With().Str("request_id", requestID).Logger()
	}
	return log.Logger
}