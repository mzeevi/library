package api

import (
	"context"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/data"
	"github.com/mzeevi/library/internal/query"
	"time"
)

const (
	errPositiveIntegerMsg       = "%s must be a positive integer"
	errMinLengthMsg             = "%s must have a minimum length of 1"
	errExactLengthMsg           = "%s must be exactly 13 characters long"
	errPositiveIntegerOrZeroMsg = "%s must be a positive integer or zero"
	errMinMaxGreaterMsg         = "%s cannot be greater than %s"
	errMinMaxLaterMsg           = "%s cannot be later than %s"
	errAtLeastOneItemMsg        = "%s must have at least one item"
	errMustEqualOneOfMsg        = "%s must be equal to %s or %s"
)

type SearchBookInput struct {
	GetBooksInput
	MinPages          *int       `json:"min_pages,omitempty"`
	MaxPages          *int       `json:"max_pages,omitempty"`
	MinEdition        *int       `json:"min_edition,omitempty"`
	MaxEdition        *int       `json:"max_edition,omitempty"`
	MinPublishedAt    *time.Time `json:"min_published_at,omitempty"`
	MaxPublishedAt    *time.Time `json:"max_published_at,omitempty"`
	Title             *string    `json:"title,omitempty"`
	ISBN              *string    `json:"isbn,omitempty"`
	Authors           []string   `json:"authors,omitempty"`
	Publishers        []string   `json:"publishers,omitempty"`
	Genres            []string   `json:"genres,omitempty"`
	MinCopies         *int       `json:"min_copies,omitempty"`
	MaxCopies         *int       `json:"max_copies,omitempty"`
	MinBorrowedCopies *int       `json:"min_borrowed_copies,omitempty"`
	MaxBorrowedCopies *int       `json:"max_borrowed_copies,omitempty"`
}

type SearchBooksOutput struct {
	Body BooksInfo
}

type SearchPatronsInput struct {
	GetPatronsInput
	Category *string `json:"category,omitempty"`
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
}

type SearchPatronsOutput struct {
	Body PatronsInfo
}

type SearchTransactionsInput struct {
	GetTransactionsInput
	PatronID      *string    `json:"patron_id,omitempty"`
	BookID        *string    `json:"book_id,omitempty"`
	Status        *string    `json:"status,omitempty"`
	MinBorrowedAt *time.Time `json:"min_borrowed_at,omitempty"`
	MaxBorrowedAt *time.Time `json:"max_borrowed_at,omitempty"`
	MinDueDate    *time.Time `json:"min_due_date,omitempty"`
	MaxDueDate    *time.Time `json:"max_due_date,omitempty"`
	MinReturnedAt *time.Time `json:"min_returned_at,omitempty"`
	MaxReturnedAt *time.Time `json:"max_returned_at,omitempty"`
	MinCreatedAt  *time.Time `json:"min_created_at,omitempty"`
	MaxCreatedAt  *time.Time `json:"max_created_at,omitempty"`
}

type SearchTransactionsOutput struct {
	Body TransactionsInfo
}

// Resolve validates the input in SearchPatronsInput.
func (s *SearchPatronsInput) Resolve(ctx huma.Context) []error {
	var errs []error

	if name, err := query.ResolveString(ctx, query.NameKey); err != nil {
		errs = append(errs, err)
	} else {
		s.Name = name
	}

	if s.Name != nil {
		if len(*s.Name) < 1 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.NameKey),
				Message:  fmt.Sprintf(errMinLengthMsg, query.NameKey),
				Value:    *s.Name,
			})
		}
	}

	if email, err := query.ResolveString(ctx, query.EmailKey); err != nil {
		errs = append(errs, err)
	} else {
		s.Email = email
	}

	if s.Email != nil {
		errs = append(errs, validateEmail(s.Email, "body.email"))
	}

	if category, err := query.ResolveString(ctx, query.CategoryKey); err != nil {
		errs = append(errs, err)
	} else {
		s.Category = category
	}

	if s.Category != nil {
		if *s.Category != studentCategory && *s.Category != teacherCategory {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.CategoryKey),
				Message:  fmt.Sprintf(errMustEqualOneOfMsg, query.CategoryKey, studentCategory, teacherCategory),
				Value:    *s.Category,
			})
		}
	}

	return errs
}

// Resolve validates the input in SearchTransactionsInput.
func (s *SearchTransactionsInput) Resolve(ctx huma.Context) []error {
	var errs []error

	if patronID, err := query.ResolveString(ctx, query.PatronIDKey); err != nil {
		errs = append(errs, err)
	} else {
		s.PatronID = patronID
	}
	if s.PatronID != nil && len(*s.PatronID) < 1 {
		errs = append(errs, &huma.ErrorDetail{
			Location: fmt.Sprintf("%s.%s", query.Key, query.PatronIDKey),
			Message:  fmt.Sprintf(errMinLengthMsg, query.PatronIDKey),
			Value:    *s.PatronID,
		})
	}

	if bookID, err := query.ResolveString(ctx, query.BookIDKey); err != nil {
		errs = append(errs, err)
	} else {
		s.BookID = bookID
	}
	if s.BookID != nil && len(*s.BookID) < 1 {
		errs = append(errs, &huma.ErrorDetail{
			Location: fmt.Sprintf("%s.%s", query.Key, query.BookIDKey),
			Message:  fmt.Sprintf(errMinLengthMsg, query.BookIDKey),
			Value:    *s.BookID,
		})
	}

	if status, err := query.ResolveString(ctx, query.StatusKey); err != nil {
		errs = append(errs, err)
	} else {
		s.Status = status
	}
	if s.Status != nil {
		if *s.Status != string(data.TransactionStatusReturned) && *s.Status != string(data.TransactionStatusBorrowed) {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.StatusKey),
				Message:  fmt.Sprintf(errMustEqualOneOfMsg, query.StatusKey, data.TransactionStatusReturned, data.TransactionStatusBorrowed),
				Value:    *s.Status,
			})
		}
	}

	if minBorrowedAt, err := query.ResolveTime(ctx, query.MinBorrowedAtKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinBorrowedAt = minBorrowedAt
	}

	if maxBorrowedAt, err := query.ResolveTime(ctx, query.MaxBorrowedAtKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxBorrowedAt = maxBorrowedAt
	}
	if s.MinBorrowedAt != nil && s.MaxBorrowedAt != nil {
		if s.MinBorrowedAt.After(*s.MaxBorrowedAt) {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s, %s.%s", query.Key, query.MinBorrowedAtKey, query.Key, query.MaxBorrowedAtKey),
				Message:  fmt.Sprintf(errMinMaxLaterMsg, query.MinBorrowedAtKey, query.MaxBorrowedAtKey),
				Value:    s.MinBorrowedAt,
			})
		}
	}

	if minDueDate, err := query.ResolveTime(ctx, query.MinDueDateKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinDueDate = minDueDate
	}

	if maxDueDate, err := query.ResolveTime(ctx, query.MaxDueDateKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxDueDate = maxDueDate
	}
	if s.MinDueDate != nil && s.MaxDueDate != nil {
		if s.MinDueDate.After(*s.MaxDueDate) {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s, %s.%s", query.Key, query.MinDueDateKey, query.Key, query.MaxDueDateKey),
				Message:  fmt.Sprintf(errMinMaxLaterMsg, query.MinDueDateKey, query.MaxDueDateKey),
				Value:    s.MinDueDate,
			})
		}
	}

	if minReturnedAt, err := query.ResolveTime(ctx, query.MinReturnedAtKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinReturnedAt = minReturnedAt
	}

	if maxReturnedAt, err := query.ResolveTime(ctx, query.MaxReturnedAtKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxReturnedAt = maxReturnedAt
	}

	if minCreatedAt, err := query.ResolveTime(ctx, query.MinCreatedAtKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinCreatedAt = minCreatedAt
	}

	if maxCreatedAt, err := query.ResolveTime(ctx, query.MaxCreatedAtKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxCreatedAt = maxCreatedAt
	}
	if s.MinCreatedAt != nil && s.MaxCreatedAt != nil {
		if s.MinCreatedAt.After(*s.MaxCreatedAt) {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s, %s.%s", query.Key, query.MinCreatedAtKey, query.Key, query.MaxCreatedAtKey),
				Message:  fmt.Sprintf(errMinMaxLaterMsg, query.MinCreatedAtKey, query.MaxCreatedAtKey),
				Value:    s.MinCreatedAt,
			})
		}
	}

	return errs
}

// Resolve validates the input in SearchBookInput.
func (s *SearchBookInput) Resolve(ctx huma.Context) []error {
	var errs []error

	if minPages, err := query.ResolveInt(ctx, query.MinPagesKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinPages = minPages
	}

	if maxPages, err := query.ResolveInt(ctx, query.MaxPagesKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxPages = maxPages
	}

	if s.MinPages != nil {
		if *s.MinPages <= 0 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.MinPagesKey),
				Message:  fmt.Sprintf(errPositiveIntegerMsg, query.MinPagesKey),
				Value:    *s.MinPages,
			})
		}
	}

	if s.MaxPages != nil {
		if *s.MaxPages <= 0 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.MaxPagesKey),
				Message:  fmt.Sprintf(errPositiveIntegerMsg, query.MaxPagesKey),
				Value:    *s.MaxPages,
			})
		}
	}
	if s.MinPages != nil && s.MaxPages != nil {
		if *s.MinPages > *s.MaxPages {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s, %s.%s", query.Key, query.MinPagesKey, query.Key, query.MaxPagesKey),
				Message:  fmt.Sprintf(errMinMaxGreaterMsg, query.MinPagesKey, query.MaxPagesKey),
				Value:    *s.MinPages,
			})
		}
	}

	if minEdition, err := query.ResolveInt(ctx, query.MinEditionKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinEdition = minEdition
	}

	if maxEdition, err := query.ResolveInt(ctx, query.MaxEditionKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxEdition = maxEdition
	}

	if s.MinEdition != nil {
		if *s.MinEdition <= 0 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.MinEditionKey),
				Message:  fmt.Sprintf(errPositiveIntegerMsg, query.MinEditionKey),
				Value:    *s.MinEdition,
			})
		}
	}

	if s.MaxEdition != nil {
		if *s.MaxEdition <= 0 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.MaxEditionKey),
				Message:  fmt.Sprintf(errPositiveIntegerMsg, query.MaxEditionKey),
				Value:    *s.MaxEdition,
			})
		}
	}
	if s.MinEdition != nil && s.MaxEdition != nil {
		if *s.MinEdition > *s.MaxEdition {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s, %s.%s", query.Key, query.MinEditionKey, query.Key, query.MaxEditionKey),
				Message:  fmt.Sprintf(errMinMaxGreaterMsg, query.MinEditionKey, query.MaxEditionKey),
				Value:    *s.MinEdition,
			})
		}
	}

	if minCopies, err := query.ResolveInt(ctx, query.MinCopiesKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinCopies = minCopies
	}

	if maxCopies, err := query.ResolveInt(ctx, query.MaxCopiesKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxCopies = maxCopies
	}

	if s.MinCopies != nil {
		if *s.MinCopies <= 0 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.MinCopiesKey),
				Message:  fmt.Sprintf(errPositiveIntegerMsg, query.MinCopiesKey),
				Value:    *s.MinCopies,
			})
		}
	}

	if s.MaxCopies != nil {
		if *s.MaxCopies <= 0 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.MaxCopiesKey),
				Message:  fmt.Sprintf(errPositiveIntegerMsg, query.MaxCopiesKey),
				Value:    *s.MaxCopies,
			})
		}
	}
	if s.MinCopies != nil && s.MaxCopies != nil {
		if *s.MinCopies > *s.MaxCopies {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s, %s.%s", query.Key, query.MinCopiesKey, query.Key, query.MaxCopiesKey),
				Message:  fmt.Sprintf(errMinMaxGreaterMsg, query.MinCopiesKey, query.MaxCopiesKey),
				Value:    *s.MinCopies,
			})
		}
	}

	if minBorrowedCopies, err := query.ResolveInt(ctx, query.MinBorrowedCopiesKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinBorrowedCopies = minBorrowedCopies
	}

	if maxBorrowedCopies, err := query.ResolveInt(ctx, query.MaxBorrowedCopiesKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxBorrowedCopies = maxBorrowedCopies
	}

	if s.MinBorrowedCopies != nil {
		if *s.MinBorrowedCopies < 0 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.MinBorrowedCopiesKey),
				Message:  fmt.Sprintf(errPositiveIntegerOrZeroMsg, query.MinBorrowedCopiesKey),
				Value:    *s.MinBorrowedCopies,
			})
		}
	}

	if s.MaxBorrowedCopies != nil {
		if *s.MaxBorrowedCopies <= 0 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.MaxBorrowedCopiesKey),
				Message:  fmt.Sprintf(errPositiveIntegerMsg, query.MaxBorrowedCopiesKey),
				Value:    *s.MaxBorrowedCopies,
			})
		}
	}
	if s.MinBorrowedCopies != nil && s.MaxBorrowedCopies != nil {
		if *s.MinBorrowedCopies > *s.MaxBorrowedCopies {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s, %s.%s", query.Key, query.MinBorrowedCopiesKey, query.Key, query.MaxBorrowedCopiesKey),
				Message:  fmt.Sprintf(errMinMaxGreaterMsg, query.MinBorrowedCopiesKey, query.MaxBorrowedCopiesKey),
				Value:    *s.MinBorrowedCopies,
			})
		}
	}

	if title, err := query.ResolveString(ctx, query.TitleKey); err != nil {
		errs = append(errs, err)
	} else {
		s.Title = title
	}

	if s.Title != nil {
		if len(*s.Title) < 1 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.TitleKey),
				Message:  fmt.Sprintf(errMinLengthMsg, query.TitleKey),
				Value:    *s.Title,
			})
		}
	}

	if isbn, err := query.ResolveString(ctx, query.ISBNKey); err != nil {
		errs = append(errs, err)
	} else {
		s.ISBN = isbn
	}

	if s.ISBN != nil {
		if len(*s.ISBN) != 13 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.ISBNKey),
				Message:  fmt.Sprintf(errExactLengthMsg, query.ISBNKey),
				Value:    *s.ISBN,
			})
		}
	}

	if authors, err := query.ResolveStringSlice(ctx, query.AuthorsKey); err != nil {
		errs = append(errs, err)
	} else {
		s.Authors = authors
	}

	if s.Authors != nil {
		if len(s.Authors) < 1 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.AuthorsKey),
				Message:  fmt.Sprintf(errAtLeastOneItemMsg, query.AuthorsKey),
				Value:    s.Authors,
			})
		}
	}

	if publishers, err := query.ResolveStringSlice(ctx, query.PublishersKey); err != nil {
		errs = append(errs, err)
	} else {
		s.Publishers = publishers
	}

	if s.Publishers != nil {
		if len(s.Publishers) < 1 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.PublishersKey),
				Message:  fmt.Sprintf(errAtLeastOneItemMsg, query.PublishersKey),
				Value:    s.Publishers,
			})
		}
	}

	if genres, err := query.ResolveStringSlice(ctx, query.PublishersKey); err != nil {
		errs = append(errs, err)
	} else {
		s.Genres = genres
	}

	if s.Genres != nil {
		if len(s.Genres) < 1 {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s", query.Key, query.GenresKey),
				Message:  fmt.Sprintf(errAtLeastOneItemMsg, query.GenresKey),
				Value:    s.Genres,
			})
		}
	}

	if minPublishedAt, err := query.ResolveTime(ctx, query.MinPublishedAtKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MinPublishedAt = minPublishedAt
	}
	if maxPublishedAt, err := query.ResolveTime(ctx, query.MaxPublishedAtKey); err != nil {
		errs = append(errs, err)
	} else {
		s.MaxPublishedAt = maxPublishedAt
	}

	if s.MinPublishedAt != nil && s.MaxPublishedAt != nil {
		if s.MinPublishedAt.After(*s.MaxPublishedAt) {
			errs = append(errs, &huma.ErrorDetail{
				Location: fmt.Sprintf("%s.%s %s.%s", query.Key, query.MinPublishedAtKey, query.Key, query.MaxPublishedAtKey),
				Message:  fmt.Sprintf(errMinMaxLaterMsg, query.MinPublishedAtKey, query.MaxPublishedAtKey),
				Value:    s.MinPublishedAt.Format(time.RFC3339),
			})
		}
	}

	return errs
}

// searchBookHandler handles the search for books based on the provided input filters and pagination.
func (app *Application) searchBookHandler(ctx context.Context, input *SearchBookInput) (*SearchBooksOutput, error) {
	paginator := data.Paginator{Page: input.Page, PageSize: input.PageSize}
	filter := data.BookFilter{}

	if input.MinPages != nil {
		filter.MinPages = input.MinPages
	}
	if input.MaxPages != nil {
		filter.MaxPages = input.MaxPages
	}
	if input.MinEdition != nil {
		filter.MinEdition = input.MinEdition
	}
	if input.MaxEdition != nil {
		filter.MaxEdition = input.MaxEdition
	}
	if input.MinCopies != nil {
		filter.MinCopies = input.MinCopies
	}
	if input.MaxCopies != nil {
		filter.MaxCopies = input.MaxCopies
	}
	if input.MinBorrowedCopies != nil {
		filter.MinBorrowedCopies = input.MinBorrowedCopies
	}
	if input.MaxBorrowedCopies != nil {
		filter.MaxBorrowedCopies = input.MaxBorrowedCopies
	}
	if input.Title != nil {
		filter.Title = input.Title
	}
	if input.ISBN != nil {
		filter.ISBN = input.ISBN
	}

	if input.Authors != nil {
		filter.Authors = input.Authors
	}
	if input.Publishers != nil {
		filter.Publishers = input.Publishers
	}
	if input.Genres != nil {
		filter.Genres = input.Genres
	}

	if input.MinPublishedAt != nil {
		filter.MinPublishedAt = input.MinPublishedAt
	}
	if input.MaxPublishedAt != nil {
		filter.MaxPublishedAt = input.MaxPublishedAt
	}

	sorter := data.Sorter{Field: input.Sort, SortSafelist: supportedBooksSortFields}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	books, metadata, err := app.Models.Books.GetAll(ctx, filter, paginator, sorter)
	if err != nil {
		return &SearchBooksOutput{}, err
	}

	resp := &SearchBooksOutput{
		Body: BooksInfo{
			Books:    books,
			Metadata: metadata,
		},
	}

	return resp, nil
}

// searchPatronsHandler handles the search for patrons based on the provided input filters and pagination.
func (app *Application) searchPatronsHandler(ctx context.Context, input *SearchPatronsInput) (*SearchPatronsOutput, error) {
	paginator := data.Paginator{Page: input.Page, PageSize: input.PageSize}
	filter := data.PatronFilter{}

	if input.Name != nil {
		filter.Name = input.Name
	}

	if input.Email != nil {
		filter.Email = input.Email
	}

	if input.Category != nil {
		filter.Category = input.Category
	}

	sorter := data.Sorter{Field: input.Sort, SortSafelist: supportedPatronsSortFields}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	patrons, metadata, err := app.Models.Patrons.GetAll(ctx, filter, paginator, sorter)
	if err != nil {
		return &SearchPatronsOutput{}, err
	}

	resp := &SearchPatronsOutput{
		Body: PatronsInfo{
			Patrons:  patrons,
			Metadata: metadata,
		},
	}

	return resp, nil
}

// searchTransactionsHandler handles the search for transactions based on the provided input filters and pagination.
func (app *Application) searchTransactionsHandler(ctx context.Context, input *SearchTransactionsInput) (*SearchTransactionsOutput, error) {
	paginator := data.Paginator{Page: input.Page, PageSize: input.PageSize}
	filter := data.TransactionFilter{}

	if input.PatronID != nil {
		filter.PatronID = input.PatronID
	}
	if input.BookID != nil {
		filter.BookID = input.BookID
	}

	if input.Status != nil {
		filter.Status = input.Status
	}

	if input.MinBorrowedAt != nil {
		filter.MinBorrowedAt = input.MinBorrowedAt
	}
	if input.MaxBorrowedAt != nil {
		filter.MaxBorrowedAt = input.MaxBorrowedAt
	}

	if input.MinDueDate != nil {
		filter.MinDueDate = input.MinDueDate
	}
	if input.MaxDueDate != nil {
		filter.MaxDueDate = input.MaxDueDate
	}

	if input.MinReturnedAt != nil {
		filter.MinReturnedAt = input.MinReturnedAt
	}
	if input.MaxReturnedAt != nil {
		filter.MaxReturnedAt = input.MaxReturnedAt
	}

	if input.MinCreatedAt != nil {
		filter.MinCreatedAt = input.MinCreatedAt
	}
	if input.MaxCreatedAt != nil {
		filter.MaxCreatedAt = input.MaxCreatedAt
	}

	if input.MinReturnedAt != nil {
		filter.MinReturnedAt = input.MinReturnedAt
	}
	if input.MaxReturnedAt != nil {
		filter.MaxReturnedAt = input.MaxReturnedAt
	}

	sorter := data.Sorter{Field: input.Sort, SortSafelist: supportedTransactionsSortFields}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	transactions, metadata, err := app.Models.Transactions.GetAll(ctx, filter, paginator, sorter)
	if err != nil {
		return &SearchTransactionsOutput{}, err
	}

	resp := &SearchTransactionsOutput{
		Body: TransactionsInfo{
			Transactions: transactions,
			Metadata:     metadata,
		},
	}

	return resp, nil
}
