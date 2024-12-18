package api

import (
	"time"
)

const (
	studentCategory = "student"
	teacherCategory = "teacher"
)

const (
	errNotFoundMsg = "the requested resource could not be found"
	errConflictMsg = "unable to update the record due to an edit conflict, please try again"
)

const (
	queryKey = "query"

	minPagesQuery          = "min_pages"
	maxPagesQuery          = "max_pages"
	minEditionQuery        = "min_edition"
	maxEditionQuery        = "max_edition"
	minPublishedAtQuery    = "min_published_at"
	maxPublishedAtQuery    = "max_published_at"
	titleQuery             = "title"
	isbnQuery              = "isbn"
	authorsQuery           = "authors"
	publishersQuery        = "publishers"
	genresQuery            = "genres"
	minCopiesQuery         = "min_copies"
	maxCopiesQuery         = "max_copies"
	minBorrowedCopiesQuery = "min_borrowed_copies"
	maxBorrowedCopiesQuery = "max_borrowed_copies"

	patronIDQuery      = "patron_id"
	bookIDQuery        = "book_id"
	statusQuery        = "status"
	minBorrowedAtQuery = "min_borrowed_at"
	maxBorrowedAtQuery = "max_borrowed_at"
	minDueDateQuery    = "min_due_date"
	maxDueDateQuery    = "max_due_date"
	minReturnedAtQuery = "min_returned_at"
	maxReturnedAtQuery = "max_returned_at"
	minCreatedAtQuery  = "min_created_at"
	maxCreatedAtQuery  = "max_created_at"

	categoryQuery = "category"
	nameQuery     = "name"
	emailQuery    = "email"
)

var (
	timeout = 10 * time.Second

	emailRX = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

	supportedBooksSortFields = []string{
		"id", "pages", "edition", "copies", "borrowedCopies", "publishedAt", "title", "isbn",
		"-id", "-pages", "-edition", "-copies", "-borrowedCopies", "-publishedAt", "-title", "-isbn",
	}

	supportedPatronsSortFields = []string{
		"category", "name", "email",
		"-category", "-name", "-email",
	}

	supportedTransactionsSortFields = []string{
		"patronID", "bookID", "status", "borrowed_at", "due_date", "returned_at",
		"-patronID", "-bookID", "-status", "-borrowed_at", "-due_date", "-returned_at",
	}
)

type PaginationInput struct {
	Page     int64 `json:"page" query:"page" minimum:"1" maximum:"1000" default:"1"`
	PageSize int64 `json:"pageSize" query:"pageSize" minimum:"1" maximum:"1000" default:"10"`
}
