package api

import (
	"time"
)

const (
	errNotFoundMsg        = "the requested resource could not be found"
	errConflictMsg        = "unable to update the record due to an edit conflict, please try again"
	errIDAlreadyExistsMsg = "a resource with this ID address already exists"
)

const (
	studentCategory = "student"
	teacherCategory = "teacher"
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
