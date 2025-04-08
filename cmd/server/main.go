package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jamesmeyerr/credit-card-validator/internal/api"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up routes
	http.HandleFunc("/validate", api.ValidationHandler)

	// Start server
	serverAddr := fmt.Sprintf(":%s", port)
	fmt.Printf("Server running on http://localhost%s\n", serverAddr)
	fmt.Println("Try validating a credit card with: curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"4532015112830366\"}' http://localhost:8080/validate")
	
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}