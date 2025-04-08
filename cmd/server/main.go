package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
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

	// Create middleware components
	rateLimiter := middleware.NewRateLimiter(RateLimit, BucketSize, CleanupInterval)
	defer rateLimiter.Shutdown()
	
	sanitizer := middleware.NewInputSanitizer(middleware.DefaultSanitizationConfig())

	// Create router
	mux := http.NewServeMux()
	
	// API endpoint
	mux.HandleFunc("/validate", api.ValidationHandler)

	// Static file server for web frontend
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve the main HTML page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join("web", "templates", "index.html"))
	})

	// Build the middleware chain - order matters:
	// 1. Logging (outermost) - captures all requests
	// 2. Rate limiting - prevents abuse
	// 3. Request sanitization - cleans inputs before processing
	
	// Apply middleware chain - logging applies to everything
	handler := middleware.LoggingMiddleware(mux)
	
	// For the validate endpoint, add sanitization
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/validate" {
			sanitizer.SanitizeMiddleware(http.HandlerFunc(api.ValidationHandler)).ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	})
	
	// Rate limiting is the final layer
	handler = rateLimiter.RateLimitMiddleware(apiHandler)

	// Create server with all middleware applied
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Channel for graceful shutdown signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a separate goroutine
	go func() {
		log.Info().
			Str("port", port).
			Float64("rate_limit", RateLimit*60).
			Int("burst_size", BucketSize).
			Bool("sanitization", true).
			Msg("Starting Credit Card Validation Service")

		fmt.Printf("Credit Card Validation Service\n")
		fmt.Printf("==============================\n")
		fmt.Printf("Server running on http://localhost%s\n", server.Addr)
		fmt.Printf("Web interface: http://localhost%s\n", server.Addr)
		fmt.Printf("API endpoint: http://localhost%s/validate\n", server.Addr)
		fmt.Printf("Rate limit: %.1f requests per minute per IP (max burst: %d)\n", RateLimit*60, BucketSize)
		fmt.Printf("Input sanitization: Enabled\n")
		fmt.Printf("Structured logging: Enabled\n")
		fmt.Printf("==============================\n")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interruption signal
	<-done
	log.Info().Msg("Shutting down server...")

	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}
	
	log.Info().Msg("Server gracefully stopped")
}