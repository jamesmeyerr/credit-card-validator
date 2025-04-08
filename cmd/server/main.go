package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jamesmeyerr/credit-card-validator/internal/api"
	"github.com/jamesmeyerr/credit-card-validator/internal/middleware"
)

// Configuration constants
const (
	// Rate limiting: 10 requests per minute per IP
	RateLimit       = 10.0 / 60.0 // tokens per second
	BucketSize      = 5           // maximum burst
	CleanupInterval = 10 * time.Minute
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a rate limiter
	rateLimiter := middleware.NewRateLimiter(RateLimit, BucketSize, CleanupInterval)
	defer rateLimiter.Shutdown()

	// Create input sanitizer with default config
	sanitizer := middleware.NewInputSanitizer(middleware.DefaultSanitizationConfig())

	// Create router
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", api.ValidationHandler)

	// Apply middleware (order matters - sanitization first, then rate limiting)
	sanitizedHandler := sanitizer.SanitizeMiddleware(mux)
	rateLimitedHandler := rateLimiter.RateLimitMiddleware(sanitizedHandler)

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      rateLimitedHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Channel for graceful shutdown signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a separate goroutine
	go func() {
		fmt.Printf("Credit Card Validation Service\n")
		fmt.Printf("==============================\n")
		fmt.Printf("Server running on http://localhost%s\n", server.Addr)
		fmt.Printf("Rate limit: %.1f requests per minute per IP (max burst: %d)\n", RateLimit*60, BucketSize)
		fmt.Printf("Input sanitization: Enabled\n")
		fmt.Printf("\nAPI Examples:\n")
		fmt.Printf("- Basic validation: curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"4532015112830366\"}' http://localhost%s/validate\n", port)
		fmt.Printf("- With expiry date: curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"5555555555554444\",\"expiry_date\":\"12/25\"}' http://localhost%s/validate\n", port)
		fmt.Printf("- Complete check: curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"378282246310005\",\"expiry_date\":\"12/25\",\"cvv\":\"1234\"}' http://localhost%s/validate\n", port)
		fmt.Printf("- Visa example: curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"4111111111111111\",\"expiry_date\":\"12/25\",\"cvv\":\"123\"}' http://localhost%s/validate\n", port)
		fmt.Printf("\nSupported card networks: Visa, Mastercard, American Express, Discover, JCB, UnionPay, Diners Club, RuPay, Maestro\n")
		fmt.Printf("Note: American Express requires 4-digit CVV, all other cards use 3-digit CVV\n")
		fmt.Printf("==============================\n")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interruption signal
	<-done
	fmt.Println("\nShutting down server...")

	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	fmt.Println("Server gracefully stopped")
}