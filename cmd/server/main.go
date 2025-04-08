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
	fmt.Printf("Credit Card Validation Service\n")
	fmt.Printf("==============================\n")
	fmt.Printf("Server running on http://localhost%s\n", serverAddr)
	fmt.Printf("\nAPI Examples:\n")
	fmt.Printf("- Visa card:         curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"4532015112830366\"}' http://localhost%s/validate\n", port)
	fmt.Printf("- Mastercard:        curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"5555555555554444\"}' http://localhost%s/validate\n", port)
	fmt.Printf("- American Express:  curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"378282246310005\"}' http://localhost%s/validate\n", port)
	fmt.Printf("\nSupported card networks: Visa, Mastercard, American Express, Discover, JCB, UnionPay, Diners Club, RuPay, Maestro\n")
	fmt.Printf("==============================\n")
	
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}