package luhn

import (
	"regexp"
	"strings"
)

// CardInfo contains validation results and card network information
type CardInfo struct {
	Valid      bool   `json:"valid"`
	Network    string `json:"network,omitempty"`
	CardLength int    `json:"card_length,omitempty"`
}

// ValidateCard checks if a credit card number is valid and identifies the network
func ValidateCard(cardNumber string) CardInfo {
	// Remove any spaces or dashes
	cleanedNumber := cleanCardNumber(cardNumber)

	// Create response object
	result := CardInfo{
		Valid:      false,
		CardLength: len(cleanedNumber),
	}

	// Skip validation if length is too short
	if len(cleanedNumber) < 2 {
		return result
	}

	// Check if the number passes the Luhn algorithm
	result.Valid = isLuhnValid(cleanedNumber)
	
	// Identify the card network
	result.Network = identifyCardNetwork(cleanedNumber)

	return result
}

// cleanCardNumber removes any non-digit characters
func cleanCardNumber(cardNumber string) string {
	var cleaned strings.Builder
	for _, r := range cardNumber {
		if r >= '0' && r <= '9' {
			cleaned.WriteRune(r)
		}
	}
	return cleaned.String()
}

// isLuhnValid implements the Luhn algorithm to validate card numbers
func isLuhnValid(cardNumber string) bool {
	var digits []int
	for _, r := range cardNumber {
		digits = append(digits, int(r-'0'))
	}

	// Check if we have a valid number of digits
	if len(digits) < 2 {
		return false
	}

	// Starting from the rightmost digit and moving left
	// double the value of every second digit
	sum := 0
	parity := len(digits) % 2
	
	for i, digit := range digits {
		// Double every second digit, starting from the right
		if i%2 == parity {
			digit *= 2
			// If doubling results in a number greater than 9, subtract 9
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	// If the total modulo 10 is 0, then the number is valid
	return sum%10 == 0
}

// identifyCardNetwork determines the payment network based on card prefix and length
func identifyCardNetwork(cardNumber string) string {
	// Common patterns for major card networks
	patterns := map[string]string{
		// Visa: Starts with 4, length 13, 16, or 19
		`^4\d{12}(?:\d{3})?(?:\d{3})?$`: "Visa",

		// Mastercard: Starts with 51-55 or 2221-2720, length 16
		`^5[1-5]\d{14}$`:     "Mastercard",
		`^2(?:2(?:2[1-9]|[3-9]\d)|[3-6]\d{2}|7(?:[01]\d|20))\d{12}$`: "Mastercard",

		// American Express: Starts with 34 or 37, length 15
		`^3[47]\d{13}$`: "American Express",

		// Discover: Starts with 6011, 622126-622925, 644-649, 65, length 16-19
		`^6(?:011|5\d{2})\d{12,15}$`: "Discover",
		`^6(?:44|45|46|47|48|49)\d{13,16}$`: "Discover",
		`^6(?:22(?:12[6-9]|1[3-9]\d|[2-9]\d{2})|2[3-9]\d{2}|[3-9]\d{3})\d{10,13}$`: "Discover",

		// JCB: Starts with 3528-3589, length 16-19
		`^35(?:2[89]|[3-8]\d)\d{12,15}$`: "JCB",

		// UnionPay: Starts with 62, length 16-19
		`^62\d{14,17}$`: "UnionPay",

		// Diners Club: Starts with 300-305, 36, 38-39, length 14-19
		`^3(?:0[0-5]|[68]\d)\d{11,16}$`: "Diners Club",

		// RuPay: Starts with 60, 6521, 6522, length 16
		`^60\d{14}$`: "RuPay",
		`^652[12]\d{13}$`: "RuPay",
		
		// Maestro: Starts with 5018, 5020, 5038, 5893, 6304, 6759, 6761, 6762, 6763, length 16-19
		`^(?:5(?:018|0[23]8|[68]93)|6(?:304|759|7(?:6[1-3])))\d{10,13}$`: "Maestro",
	}

	for pattern, network := range patterns {
		match, _ := regexp.MatchString(pattern, cardNumber)
		if match {
			return network
		}
	}

	return "Unknown"
}