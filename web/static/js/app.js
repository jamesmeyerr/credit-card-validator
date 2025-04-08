document.addEventListener('DOMContentLoaded', function() {
    const cardForm = document.getElementById('card-form');
    const cardNumberInput = document.getElementById('card-number');
    const expiryDateInput = document.getElementById('expiry-date');
    const cvvInput = document.getElementById('cvv');
    const resultDiv = document.getElementById('result');
    const resultValid = document.getElementById('result-valid');
    const resultInvalid = document.getElementById('result-invalid');
    const validMessage = document.getElementById('valid-message');
    const invalidMessage = document.getElementById('invalid-message');
    const resultDetails = document.getElementById('result-details');
    const cardTypeDisplay = document.getElementById('card-type');

    // Credit card network icons
    const cardIcons = {
        'visa': document.getElementById('visa-icon'),
        'mastercard': document.getElementById('mastercard-icon'),
        'amex': document.getElementById('amex-icon'),
        'discover': document.getElementById('discover-icon'),
        'generic': document.getElementById('generic-icon')
    };

    // Format card number as it's typed
    cardNumberInput.addEventListener('input', function(e) {
        // Remove non-digit characters
        let value = this.value.replace(/\D/g, '');
        
        // Identify card type from prefix
        identifyCardType(value);
        
        // Format with spaces
        if (value.length > 0) {
            // Different formatting for Amex (4-6-5 pattern)
            if (isAmex(value)) {
                value = value.replace(/^(\d{4})(\d{0,6})(\d{0,5}).*/, function(match, p1, p2, p3) {
                    let result = p1;
                    if (p2) result += ' ' + p2;
                    if (p3) result += ' ' + p3;
                    return result;
                });
            } else {
                // Standard 4-digit grouping
                value = value.replace(/(\d{4})(?=\d)/g, '$1 ');
            }
        }
        
        this.value = value;
    });

    // Format expiry date as MM/YY
    expiryDateInput.addEventListener('input', function(e) {
        let value = this.value.replace(/\D/g, '');
        
        if (value.length > 0) {
            if (value.length <= 2) {
                // Just the month
                value = value.replace(/^(\d{0,2})/, '$1');
            } else {
                // Month and year
                value = value.replace(/^(\d{2})(\d{0,2})/, '$1/$2');
            }
        }
        
        this.value = value;
    });

    // Restrict CVV to numbers only and set max length
    cvvInput.addEventListener('input', function(e) {
        this.value = this.value.replace(/\D/g, '');
        // Check if current card type is Amex
        if (cardTypeDisplay.textContent.includes('American Express')) {
            this.setAttribute('maxlength', '4');
        } else {
            this.setAttribute('maxlength', '3');
        }
    });

    // Identify card type based on number
    function identifyCardType(cardNumber) {
        // Reset all icons to gray
        Object.values(cardIcons).forEach(icon => {
            icon.classList.remove('text-blue-600', 'text-orange-600', 'text-blue-400', 'text-orange-500');
            icon.classList.add('text-gray-400');
        });

        // Simple regex patterns for major card types
        let cardType = '';
        
        // Visa: Starts with 4
        if (/^4/.test(cardNumber)) {
            cardType = 'Visa';
            cardIcons.visa.classList.remove('text-gray-400');
            cardIcons.visa.classList.add('text-blue-600');
        } 
        // Mastercard: Starts with 51-55 or 2221-2720
        else if (/^(5[1-5]|2[2-7][2][01])/.test(cardNumber)) {
            cardType = 'Mastercard';
            cardIcons.mastercard.classList.remove('text-gray-400');
            cardIcons.mastercard.classList.add('text-orange-600');
        } 
        // Amex: Starts with 34 or 37
        else if (/^3[47]/.test(cardNumber)) {
            cardType = 'American Express';
            cardIcons.amex.classList.remove('text-gray-400');
            cardIcons.amex.classList.add('text-blue-400');
            // Update CVV max length
            cvvInput.setAttribute('maxlength', '4');
        } 
        // Discover: Starts with 6011, 644-649, 65
        else if (/^(6011|64[4-9]|65)/.test(cardNumber)) {
            cardType = 'Discover';
            cardIcons.discover.classList.remove('text-gray-400');
            cardIcons.discover.classList.add('text-orange-500');
        } 
        // Default: Unknown
        else if (cardNumber.length > 0) {
            cardType = 'Unknown card type';
            cardIcons.generic.classList.remove('text-gray-400');
            cardIcons.generic.classList.add('text-gray-600');
        } else {
            cardType = '';
            // Keep all icons gray if no input
        }
        
        cardTypeDisplay.textContent = cardType;
    }

    // Check if it's an Amex card
    function isAmex(cardNumber) {
        return /^3[47]/.test(cardNumber);
    }

    // Form submission handler
    cardForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        // Get values
        let cardNumber = cardNumberInput.value.replace(/\D/g, '');
        let expiryDate = expiryDateInput.value;
        let cvv = cvvInput.value;
        
        // Validate form
        if (!cardNumber) {
            showError('Please enter a card number');
            return;
        }

        // Create request payload
        const payload = {
            card_number: cardNumber
        };
        
        if (expiryDate) {
            payload.expiry_date = expiryDate;
        }
        
        if (cvv) {
            payload.cvv = cvv;
        }
        
        // Show loading state
        const button = cardForm.querySelector('button[type="submit"]');
        const originalText = button.textContent;
        button.textContent = 'Validating...';
        button.disabled = true;
        
        // Send API request
        fetch('/validate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Error: ' + response.status);
            }
            return response.json();
        })
        .then(data => {
            displayResult(data);
        })
        .catch(error => {
            showError('Error validating card: ' + error.message);
        })
        .finally(() => {
            // Reset button state
            button.textContent = originalText;
            button.disabled = false;
        });
    });

    // Display validation result
    function displayResult(data) {
        resultDiv.classList.remove('hidden');
        
        if (data.valid) {
            resultValid.classList.remove('hidden');
            resultInvalid.classList.add('hidden');
            validMessage.textContent = data.message || 'Valid card';
        } else {
            resultValid.classList.add('hidden');
            resultInvalid.classList.remove('hidden');
            invalidMessage.textContent = data.message || 'Invalid card';
        }
        
        // Show additional details
        let details = [];
        if (data.network) {
            details.push(`Network: ${data.network}`);
        }
        if (data.card_length) {
            details.push(`Length: ${data.card_length} digits`);
        }
        if (data.expiry_format_ok !== undefined) {
            details.push(`Expiry format: ${data.expiry_format_ok ? 'Valid' : 'Invalid'}`);
        }
        if (data.expiry_valid !== undefined && data.expiry_format_ok) {
            details.push(`Expiry status: ${data.expiry_valid ? 'Not expired' : 'Expired'}`);
        }
        if (data.cvv_valid !== undefined) {
            details.push(`CVV: ${data.cvv_valid ? 'Valid' : 'Invalid'}`);
        }
        
        resultDetails.textContent = details.join(' | ');
    }

    // Show error message
    function showError(message) {
        resultDiv.classList.remove('hidden');
        resultValid.classList.add('hidden');
        resultInvalid.classList.remove('hidden');
        invalidMessage.textContent = message;
        resultDetails.textContent = '';
    }
});