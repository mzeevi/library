package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/data"
	"regexp"
	"time"
)

// isBookUnavailable checks if borrowing the specified number of copies would exceed the available stock.
func isBookUnavailable(book *data.Book, requestedCopies int) bool {
	return book.BorrowedCopies+requestedCopies > book.Copies
}

// processPatronTransactions returns a PatronTransactions slice.
func processPatronTransactions(transactions []data.Transaction, overdueFine float64) ([]patronTransaction, float64) {
	patronTransactions := make([]patronTransaction, 0)
	var totalFine float64

	for _, transaction := range transactions {
		pt := patronTransaction{
			Transaction: transaction,
			Fine:        calculateFine(transaction, overdueFine),
		}

		patronTransactions = append(patronTransactions, pt)
		totalFine = totalFine + pt.Fine
	}

	return patronTransactions, totalFine
}

// calculateFine calculates the fine for a transaction. It checks if it is overdue based on the due date.
// For overdue transactions, the fine is calculated by multiplying the number
// of overdue days by the specified overdue fine rate.
func calculateFine(transaction data.Transaction, overdueFine float64) (fine float64) {
	daysOverdue := time.Since(transaction.DueDate).Hours() / 24
	if daysOverdue > 0 {
		fine = daysOverdue * overdueFine
	}

	return fine
}

// validateEmail validates an email address against a predefined regular expression.
func validateEmail(email *string) error {
	if email == nil {
		return nil
	}
	matched, err := regexp.MatchString(emailRX, *email)
	if err != nil {
		return &huma.ErrorDetail{
			Location: "body.email",
			Message:  "Error parsing",
			Value:    *email,
		}
	}
	if !matched {
		return &huma.ErrorDetail{
			Location: "body.email",
			Message:  "Invalid email address",
			Value:    *email,
		}
	}
	return nil
}

// ptr is a generic helper function for creating a pointer to any type.
func ptr[T any](v T) *T {
	return &v
}
