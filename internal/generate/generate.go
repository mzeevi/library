package generate

import (
	"fmt"
	"github.com/mzeevi/library/internal/data"
	"strconv"
	"time"
)

// Books returns a slice of n new books.
func Books(n int) []data.Book {
	var books []data.Book

	for i := 0; i < n; i++ {
		s := strconv.Itoa(i)

		book := data.NewBook(fmt.Sprintf("test-book-%s", s), generateISBN(),
			[]string{fmt.Sprintf("test-author-1-%s", s), fmt.Sprintf("test-author-2-%s", s)},
			[]string{fmt.Sprintf("test-publisher-1-%s", s), fmt.Sprintf("test-publisher-2-%s", s)},
			[]string{fmt.Sprintf("test-genre-1-%s", s), fmt.Sprintf("test-genre-2-%s", s)},
			i, i, time.Date(2015, time.November, 20, 0, 0, 0, 0, time.UTC), 30*24*time.Hour,
		)

		books = append(books, book)
	}

	return books
}

// Patrons returns a slice of n new patrons.
func Patrons(n int, discounts map[data.PatronCategoryType]float64) ([]data.Patron, error) {
	var patrons []data.Patron
	var categoryType data.PatronCategoryType

	for i := 0; i < n; i++ {
		s := strconv.Itoa(i)

		switch i % 2 {
		case 0:
			categoryType = data.Student
		default:
			categoryType = data.Teacher
		}

		patron, err := data.NewPatron(fmt.Sprintf("test-%s", s), categoryType, discounts)
		if err != nil {
			return patrons, err
		}

		patrons = append(patrons, patron)
	}

	return patrons, nil
}
