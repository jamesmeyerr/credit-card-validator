package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

	// Apply middleware (order matters - sanitization first, then rate limiting)
	// Only apply to API endpoints, not static files
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/validate" {
			sanitizer.SanitizeMiddleware(http.HandlerFunc(api.ValidationHandler)).ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	})
	
	rateLimitedHandler := rateLimiter.RateLimitMiddleware(apiHandler)

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
		fmt.Printf("Web interface: http://localhost%s\n", server.Addr)
		fmt.Printf("API endpoint: http://localhost%s/validate\n", server.Addr)
		fmt.Printf("Rate limit: %.1f requests per minute per IP (max burst: %d)\n", RateLimit*60, BucketSize)
		fmt.Printf("Input sanitization: Enabled\n")
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