package data

import (
	"errors"
	"github.com/google/uuid"
	"math"
	"time"
)

const (
	errUnknownCategory = "unknown patron category"
)

const (
	Teacher PatronCategoryType = "teacher"
	Student PatronCategoryType = "student"
)

type PatronCategoryType string

type Patron struct {
	ID            uint32
	Name          string
	CreatedAt     time.Time
	BorrowedBooks map[string]bookDetails
	Category      patronCategory
}

type bookDetails struct {
	ISBN           string
	BorrowDuration time.Duration
	BorrowedAt     time.Time
}

type patronCategory interface {
	Discount() float64
}

type TeacherCategory struct {
	DiscountPercentage float64
}

type StudentCategory struct {
	DiscountPercentage float64
}

func (t TeacherCategory) Discount() float64 {
	return t.DiscountPercentage / 100
}

func (s StudentCategory) Discount() float64 {
	return s.DiscountPercentage / 100
}

// NewPatron creates a new Patron instance with the specified name, category type, and discount rates.
// It initializes the patron's borrowing history and assigns the appropriate discount category.
func NewPatron(name string, categoryType PatronCategoryType, discounts map[PatronCategoryType]float64) (Patron, error) {
	var category patronCategory

	switch categoryType {
	case Teacher:
		category = TeacherCategory{
			DiscountPercentage: discounts[Teacher],
		}
	case Student:
		category = StudentCategory{
			DiscountPercentage: discounts[Student],
		}
	default:
		return Patron{}, errors.New(errUnknownCategory)
	}

	return Patron{
		ID:            uuid.New().ID(),
		Name:          name,
		CreatedAt:     time.Now(),
		BorrowedBooks: make(map[string]bookDetails),
		Category:      category,
	}, nil
}

// UpdatePatron updates the patron's name and/or category type based on the provided parameters.
func (p *Patron) UpdatePatron(name *string, categoryType *PatronCategoryType, discounts map[PatronCategoryType]float64) error {
	if name != nil {
		p.Name = *name
	}

	if categoryType != nil {
		switch *categoryType {
		case Teacher:
			p.Category = TeacherCategory{
				DiscountPercentage: discounts[Teacher],
			}
		case Student:
			p.Category = StudentCategory{
				DiscountPercentage: discounts[Student],
			}
		default:
			return errors.New(errUnknownCategory)
		}
	}

	return nil
}

// BorrowBook allows the patron to borrow a book by title from the available list of books.
// It updates the book's status to borrowed and records the borrowing details for the patron.
func (p *Patron) BorrowBook(title string, books []Book) {
	book := getBookByTitle(title, books)
	book.markBookAsBorrowed()

	borrowed := bookDetails{
		ISBN:           book.ISBN,
		BorrowDuration: book.BorrowDuration,
		BorrowedAt:     time.Now(),
	}

	if p.BorrowedBooks == nil {
		p.BorrowedBooks = make(map[string]bookDetails)
	}

	p.BorrowedBooks[title] = borrowed
}

// ReturnBook allows the patron to return a borrowed book by title.
// It updates the book's status to available and removes the borrowing record for the patron.
func (p *Patron) ReturnBook(title string, books []Book) {
	book := getBookByTitle(title, books)
	book.markBookAsNotBorrowed()

	delete(p.BorrowedBooks, title)
}

// GetBorrowedBooks retrieves a map of the books currently borrowed by the patron and their due dates.
func (p *Patron) GetBorrowedBooks() map[string]time.Time {
	borrowed := make(map[string]time.Time)

	for title, book := range p.BorrowedBooks {
		borrowed[title] = book.BorrowedAt.Add(book.BorrowDuration)
	}

	return borrowed
}

// CalcFine calculates the total fine for overdue books borrowed by the patron.
// The fine is based on the overdue duration of each book, a specified per-day overdue fine rate,
// and the patron's category discount.
func (p *Patron) CalcFine(overdueFine float64) float64 {
	var totalFine float64

	for _, item := range p.BorrowedBooks {
		bookFine := overdueFine * float64(daysBetween(item.BorrowedAt.Add(item.BorrowDuration), time.Now()))
		totalFine = totalFine + bookFine
	}

	return totalFine * p.Category.Discount()
}

// daysBetween calculates the number of days between two time.Time values, rounding up the result.
// If the end time is before the start time, it returns 0 or handles the error.
func daysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	if duration < 0 {
		return 0
	}

	return int(math.Ceil(duration.Hours() / 24))
}
