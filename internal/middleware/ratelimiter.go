package middleware

import (
    "net/http"
    "sync"
    "time"
    "encoding/json"
)

// RateLimiter implements a token bucket rate limiting algorithm
type RateLimiter struct {
    rate       float64     // tokens per second
    bucketSize int         // maximum tokens
    clients    map[string]*bucket
    mu         sync.Mutex
    cleanup    *time.Ticker
}

// bucket represents a token bucket for a single client
type bucket struct {
    tokens     float64
    lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, bucketSize int, cleanupInterval time.Duration) *RateLimiter {
    limiter := &RateLimiter{
        rate:       rate,
        bucketSize: bucketSize,
        clients:    make(map[string]*bucket),
        cleanup:    time.NewTicker(cleanupInterval),
    }

    // Start cleanup routine to remove stale buckets
    go func() {
        for range limiter.cleanup.C {
            limiter.cleanupStale(30 * time.Minute)
        }
    }()

    return limiter
}

// cleanupStale removes buckets that haven't been used for a while
func (rl *RateLimiter) cleanupStale(maxAge time.Duration) {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    threshold := time.Now().Add(-maxAge)
    for ip, bucket := range rl.clients {
        if bucket.lastRefill.Before(threshold) {
            delete(rl.clients, ip)
        }
    }
}

// Allow checks if a request should be allowed based on the client's IP
func (rl *RateLimiter) Allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    b, exists := rl.clients[ip]
    if !exists {
        // Create a new bucket for this client
        rl.clients[ip] = &bucket{
            tokens:     float64(rl.bucketSize) - 1, // Use one token for this request
            lastRefill: time.Now(),
        }
        return true
    }

    // Calculate token refill since last request
    now := time.Now()
    elapsed := now.Sub(b.lastRefill).Seconds()
    refill := elapsed * rl.rate
    
    // Refill the bucket (up to max capacity)
    b.tokens = min(float64(rl.bucketSize), b.tokens+refill)
    b.lastRefill = now

    // Check if enough tokens
    if b.tokens >= 1.0 {
        b.tokens -= 1.0
        return true
    }

    return false
}

// Helper function for float64 minimum
func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

// Shutdown stops the cleanup ticker
func (rl *RateLimiter) Shutdown() {
    rl.cleanup.Stop()
}

// RateLimitMiddleware creates a middleware function for rate limiting
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get client IP using the logger's getClientIP function
        ip := getClientIP(r)
        if ip == "" {
            // Log this with the application logger if needed
            http.Error(w, "Unable to determine client IP", http.StatusInternalServerError)
            return
        }

        // Check if request is allowed
        if !rl.Allow(ip) {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Rate limit exceeded, please try again later",
            })
            return
        }

        // Pass to next handler if request is allowed
        next.ServeHTTP(w, r)
    })
}