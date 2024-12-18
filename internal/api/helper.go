package api

import (
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func validateEmail(email *string, location string) error {
	if email == nil {
		return nil
	}
	matched, err := regexp.MatchString(emailRX, *email)
	if err != nil {
		return &huma.ErrorDetail{
			Location: location,
			Message:  "Error parsing",
			Value:    *email,
		}
	}
	if !matched {
		return &huma.ErrorDetail{
			Location: location,
			Message:  "Invalid email address",
			Value:    *email,
		}
	}
	return nil
}

// validateEmail validates an ID.
func validateID(id *string, location string) error {
	if id == nil {
		return nil
	}

	_, err := primitive.ObjectIDFromHex(*id)
	if err != nil {
		return &huma.ErrorDetail{
			Location: location,
			Message:  "Invalid ID",
			Value:    *id,
		}
	}

	return nil
}

// validateDueDate checks if the due date is valid, ensuring it is between 1 and 14 days from today.
func validateDueDate(t *time.Time, location string) error {
	if t == nil {
		return nil
	}

	oneDayFromNow := time.Now().Add(24 * time.Hour)
	twoWeeksFromNow := time.Now().Add(14 * 24 * time.Hour)

	valid := (*t).After(oneDayFromNow) && (*t).Before(twoWeeksFromNow)
	if !valid {
		return &huma.ErrorDetail{
			Location: location,
			Message: fmt.Sprintf(
				"Due date must be at least 1 day (after %s) and no more than 14 days (before %s) from today",
				oneDayFromNow.Format(time.RFC3339),
				twoWeeksFromNow.Format(time.RFC3339),
			),
			Value: *t,
		}
	}

	return nil
}

// ptr is a generic helper function for creating a pointer to any type.
func ptr[T any](v T) *T {
	return &v
}
