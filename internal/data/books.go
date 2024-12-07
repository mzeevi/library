package data

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID             uint32
	Pages          int
	Edition        int
	Borrowed       bool
	Published      time.Time
	CreatedAt      time.Time
	BorrowDuration time.Duration
	Title          string
	ISBN           string
	Authors        []string
	Publishers     []string
	Genres         []string
	Version        int32
}

// SearchCriteria struct holds optional search filters for each field of the Book struct
type SearchCriteria struct {
	Title        *string
	ISBN         *string
	Authors      *[]string
	Publishers   *[]string
	Genres       *[]string
	Borrowed     *bool
	MinPages     *int
	MaxPages     *int
	MinEdition   *int
	MaxEdition   *int
	MinPublished *time.Time
	MaxPublished *time.Time
}

// NewBook creates a new book with the provided details.
func NewBook(title string, isbn string, authors []string, publishers []string, genres []string, pages int, edition int, published time.Time, borrowDuration time.Duration) Book {
	return Book{
		ID:             uuid.New().ID(),
		Title:          title,
		ISBN:           isbn,
		Authors:        authors,
		Publishers:     publishers,
		Genres:         genres,
		Pages:          pages,
		Edition:        edition,
		Published:      published,
		BorrowDuration: borrowDuration,
		CreatedAt:      time.Now(),
		Borrowed:       false,
	}
}

// UpdateBook updates the book's details based on the provided parameters.
func (b *Book) UpdateBook(title *string, isbn *string, authors *[]string, publishers *[]string, genres *[]string, pages *int, edition *int, published *time.Time, borrowDuration *time.Duration) {
	if title != nil {
		b.Title = *title
	}

	if isbn != nil {
		b.ISBN = *isbn
	}

	if authors != nil {
		b.Authors = *authors
	}

	if publishers != nil {
		b.Publishers = *publishers
	}

	if genres != nil {
		b.Genres = *genres
	}

	if pages != nil {
		b.Pages = *pages
	}

	if edition != nil {
		b.Edition = *edition
	}

	if published != nil {
		b.Published = *published
	}

	if borrowDuration != nil {
		b.BorrowDuration = *borrowDuration
	}
}

// markBookAsBorrowed marks the book as currently borrowed.
func (b *Book) markBookAsBorrowed() {
	b.Borrowed = true
}

// markBookAsNotBorrowed marks the book as not borrowed.
func (b *Book) markBookAsNotBorrowed() {
	b.Borrowed = false
}

// getBookByISBN retrieves a book in the given slice by its title.
func getBookByTitle(title string, books []Book) *Book {
	for _, book := range books {
		if title == book.Title {
			return &book
		}
	}

	return nil
}

// SearchBooks filters the given books slice based on the provided criteria.
func SearchBooks(books []Book, criteria SearchCriteria) []Book {
	var results []Book

	for _, book := range books {
		if matchesAllCriteria(book, criteria) {
			results = append(results, book)
		}
	}

	return results
}

// matchesAllCriteria checks if a book matches all the search criteria.
func matchesAllCriteria(book Book, criteria SearchCriteria) bool {
	if !checkTitle(book.Title, criteria.Title) {
		return false
	}
	if !checkISBN(book.ISBN, criteria.ISBN) {
		return false
	}
	if !checkStringSlice(book.Authors, criteria.Authors) {
		return false
	}
	if !checkStringSlice(book.Publishers, criteria.Publishers) {
		return false
	}
	if !checkStringSlice(book.Genres, criteria.Genres) {
		return false
	}
	if !checkBorrowed(book.Borrowed, criteria.Borrowed) {
		return false
	}
	if !checkPages(book.Pages, criteria.MinPages, criteria.MaxPages) {
		return false
	}
	if !checkEdition(book.Edition, criteria.MinEdition, criteria.MaxEdition) {
		return false
	}
	if !checkPublished(book.Published, criteria.MinPublished, criteria.MaxPublished) {
		return false
	}

	return true
}

// contains helper function checks if a slice contains a specific string
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// checkTitle checks if the book's title matches the search criteria.
func checkTitle(bookTitle string, title *string) bool {
	if title != nil && !strings.Contains(bookTitle, *title) {
		return false
	}
	return true
}

// checkISBN checks if the book's ISBN matches the search criteria.
func checkISBN(bookISBN string, isbn *string) bool {
	if isbn != nil && bookISBN != *isbn {
		return false
	}
	return true
}

// checkStringSlice checks if any element of the book's string slice matches the search criteria.
func checkStringSlice(s []string, criteriaSlice *[]string) bool {
	if criteriaSlice != nil {
		for _, criteria := range *criteriaSlice {
			if contains(s, criteria) {
				return true
			}
		}
		return false
	}
	return true
}

// checkBorrowed checks if the book's borrowed status matches the search criteria.
func checkBorrowed(bookBorrowed bool, borrowed *bool) bool {
	if borrowed != nil && bookBorrowed != *borrowed {
		return false
	}
	return true
}

// checkPages checks if the book's pages match the search criteria.
func checkPages(bookPages int, minPages, maxPages *int) bool {
	if minPages != nil && bookPages < *minPages {
		return false
	}
	if maxPages != nil && bookPages > *maxPages {
		return false
	}
	return true
}

// checkEdition checks if the book's edition matches the search criteria.
func checkEdition(bookEdition int, minEdition, maxEdition *int) bool {
	if minEdition != nil && bookEdition < *minEdition {
		return false
	}
	if maxEdition != nil && bookEdition > *maxEdition {
		return false
	}
	return true
}

// checkPublished checks if the book's published date matches the search criteria.
func checkPublished(bookPublished time.Time, minPublished, maxPublished *time.Time) bool {
	if minPublished != nil && bookPublished.Before(*minPublished) {
		return false
	}
	if maxPublished != nil && bookPublished.After(*maxPublished) {
		return false
	}
	return true
}
