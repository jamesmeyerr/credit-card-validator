package luhn

// IsValid checks if a credit card number is valid according to the Luhn algorithm
func IsValid(cardNumber string) bool {
	// Remove any spaces or dashes
	var digits []int
	for _, r := range cardNumber {
		if r >= '0' && r <= '9' {
			digits = append(digits, int(r-'0'))
		}
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