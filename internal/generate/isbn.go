package generate

import (
	"fmt"
	"math/rand"
	"time"
)

// generateISBN generates a valid ISBN-13.
func generateISBN() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	prefix := []int{9, 7, 8}

	regGroup := r.Intn(10)
	publisher := r.Intn(9000) + 1000
	titleID := r.Intn(90000) + 10000

	isbn := append(prefix, regGroup)
	isbn = append(isbn, splitDigits(publisher)...)
	isbn = append(isbn, splitDigits(titleID)...)

	checkDigit := calculateCheckDigit(isbn)
	isbn = append(isbn, checkDigit)

	return digitsToString(isbn)
}

// splitDigits splits an integer into its individual digits.
func splitDigits(num int) []int {
	var digits []int
	for num > 0 {
		digits = append([]int{num % 10}, digits...)
		num /= 10
	}
	return digits
}

// calculateCheckDigit computes the ISBN-13 check digit.
func calculateCheckDigit(digits []int) int {
	sum := 0
	for i, digit := range digits {
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	return (10 - (sum % 10)) % 10
}

// digitsToString converts a slice of integers to a string.
func digitsToString(digits []int) string {
	result := ""
	for _, digit := range digits {
		result += fmt.Sprintf("%d", digit)
	}
	return result
}
