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
	fmt.Printf("- Basic validation:  curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"4532015112830366\"}' http://localhost%s/validate\n", port)
	fmt.Printf("- With expiry date: curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"5555555555554444\",\"expiry_date\":\"12/25\"}' http://localhost%s/validate\n", port)
	fmt.Printf("- Complete check:   curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"378282246310005\",\"expiry_date\":\"12/25\",\"cvv\":\"1234\"}' http://localhost%s/validate\n", port)
	fmt.Printf("- Visa example:     curl -X GET -H \"Content-Type: application/json\" -d '{\"card_number\":\"4111111111111111\",\"expiry_date\":\"12/25\",\"cvv\":\"123\"}' http://localhost%s/validate\n", port)
	fmt.Printf("\nSupported card networks: Visa, Mastercard, American Express, Discover, JCB, UnionPay, Diners Club, RuPay, Maestro\n")
	fmt.Printf("Note: American Express requires 4-digit CVV, all other cards use 3-digit CVV\n")
	fmt.Printf("==============================\n")
	
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}