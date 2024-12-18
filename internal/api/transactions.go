package api

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/data"
)

type GetTransactionInput struct {
	ID string `json:"id" path:"id"`
}

type GetTransactionOutput struct {
	Body data.Transaction
}

type GetTransactionsInput struct {
	PaginationInput
	Sort string `json:"sort,omitempty" query:"sort" enum:"patronID,bookID,status,borrowed_at,due_date,returned_at,-patronID,-bookID,-status,-borrowed_at,-due_date,-returned_at"`
}

type GetTransactionsOutput struct {
	Body TransactionsInfo
}

type TransactionsInfo struct {
	Transactions []data.Transaction `json:"transactions"`
	Metadata     data.Metadata      `json:"metadata"`
}

type BorrowBookTransactionInput struct {
	Body struct {
		PatronID string    `json:"patron_id"`
		BookID   string    `json:"book_id"`
		DueDate  time.Time `json:"due_date" format:"date-time"`
		Copies   int       `json:"copies" minimum:"1" default:"1"`
	}
}

type BorrowBookTransactionOutput struct {
	Location string           `header:"Location"`
	Body     data.Transaction `json:"transaction"`
}

type ReturnBookTransactionInput struct {
	Body struct {
		PatronID string `json:"patron_id"`
		BookID   string `json:"book_id"`
		Copies   int    `json:"copies" minimum:"1" default:"1"`
	}
}

type ReturnBookTransactionOutput struct {
	Body string `json:"message"`
}

type UpdateTransactionInput struct {
	ID   string `json:"id" path:"id"`
	Body struct {
		DueDate *time.Time `json:"due_date" format:"date-time"`
	}
}

type UpdateTransactionOutput struct {
	Body data.Transaction `json:"transaction"`
}

type DeleteTransactionInput struct {
	ID string `json:"id" path:"id"`
}

type DeleteTransactionOutput struct {
	Body string `json:"message"`
}

// Resolve validates the input in BorrowBookTransactionInput.
func (t *BorrowBookTransactionInput) Resolve(ctx huma.Context) []error {
	return validateDueDate(&t.Body.DueDate)
}

// Resolve validates the input in UpdateTransactionInput.
func (t *UpdateTransactionInput) Resolve(ctx huma.Context) []error {
	return validateDueDate(t.Body.DueDate)
}

// validateDueDate checks if the due date is valid, ensuring it is between 1 and 14 days from today.
func validateDueDate(t *time.Time) []error {
	if t == nil {
		return nil
	}

	oneDayFromNow := time.Now().Add(24 * time.Hour)
	twoWeeksFromNow := time.Now().Add(14 * 24 * time.Hour)

	valid := (*t).After(oneDayFromNow) && (*t).Before(twoWeeksFromNow)
	if !valid {
		return []error{&huma.ErrorDetail{
			Location: "body.dueDate",
			Message: fmt.Sprintf(
				"Due date must be at least 1 day (after %s) and no more than 14 days (before %s) from today",
				oneDayFromNow.Format(time.RFC3339),
				twoWeeksFromNow.Format(time.RFC3339),
			),
			Value: *t,
		}}
	}
	return nil
}

// getTransactionHandler handles a request to fetch a single transaction by ID.
func (app *Application) getTransactionHandler(ctx context.Context, input *GetTransactionInput) (*GetTransactionOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	transaction, err := app.models.Transactions.Get(ctx, data.TransactionFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &GetTransactionOutput{}, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &GetTransactionOutput{}, err
		}
	}

	resp := &GetTransactionOutput{}
	resp.Body = *transaction

	return resp, nil
}

// getTransactionsHandler handles a request to fetch all transactions with pagination and sorting.
func (app *Application) getTransactionsHandler(ctx context.Context, input *GetTransactionsInput) (*GetTransactionsOutput, error) {
	paginator := data.Paginator{Page: input.Page, PageSize: input.PageSize}
	filter := data.TransactionFilter{}

	sorter := data.Sorter{Field: input.Sort, SortSafelist: supportedTransactionsSortFields}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	transactions, metadata, err := app.models.Transactions.GetAll(ctx, filter, paginator, sorter)
	if err != nil {
		return &GetTransactionsOutput{}, err
	}

	resp := &GetTransactionsOutput{
		Body: TransactionsInfo{
			Transactions: transactions,
			Metadata:     metadata,
		},
	}

	return resp, nil
}

func (app *Application) borrowBookTransactionHandler(ctx context.Context, input *BorrowBookTransactionInput) (*BorrowBookTransactionOutput, error) {
	dbClient := app.models.Transactions.Client

	session, err := dbClient.StartSession()
	if err != nil {
		return &BorrowBookTransactionOutput{}, err
	}
	defer session.EndSession(context.Background())

	sessionContext := mongo.NewSessionContext(ctx, session)
	if err = session.StartTransaction(); err != nil {
		return &BorrowBookTransactionOutput{}, err
	}

	book, err := app.models.Books.Get(sessionContext, data.BookFilter{ID: &input.Body.BookID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return nil, huma.Error404NotFound("the requested book resource could not be found")
		default:
			return &BorrowBookTransactionOutput{}, err
		}
	}

	patron, err := app.models.Patrons.Get(sessionContext, data.PatronFilter{ID: &input.Body.PatronID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &BorrowBookTransactionOutput{}, huma.Error404NotFound("the requested patron resource could not be found")
		default:
			return &BorrowBookTransactionOutput{}, err
		}
	}

	if isBookUnavailable(book, input.Body.Copies) {
		return nil, huma.Error409Conflict("not enough copies of the book are available for borrowing")
	}

	transaction := &data.Transaction{
		PatronID:   patron.ID,
		BookID:     book.ID,
		DueDate:    input.Body.DueDate,
		Status:     data.TransactionStatusBorrowed,
		BorrowedAt: time.Now(),
	}

	id, err := app.models.Transactions.Insert(sessionContext, transaction)
	if err != nil {
		return &BorrowBookTransactionOutput{}, err
	}

	book.BorrowedCopies = book.BorrowedCopies + input.Body.Copies
	err = app.models.Books.Update(sessionContext, data.BookFilter{ID: &book.ID}, book)
	if err != nil {
		return &BorrowBookTransactionOutput{}, err
	}

	if err = session.CommitTransaction(ctx); err != nil {
		return &BorrowBookTransactionOutput{}, err
	}

	resp := &BorrowBookTransactionOutput{
		Body:     *transaction,
		Location: fmt.Sprintf("%s/%s/%s", basePath, transactionsKey, id),
	}

	return resp, nil
}

func (app *Application) returnBookTransactionHandler(ctx context.Context, input *ReturnBookTransactionInput) (*ReturnBookTransactionOutput, error) {
	dbClient := app.models.Transactions.Client

	session, err := dbClient.StartSession()
	if err != nil {
		return &ReturnBookTransactionOutput{}, err
	}
	defer session.EndSession(context.Background())

	sessionContext := mongo.NewSessionContext(ctx, session)
	if err = session.StartTransaction(); err != nil {
		return &ReturnBookTransactionOutput{}, err
	}

	book, err := app.models.Books.Get(sessionContext, data.BookFilter{ID: &input.Body.BookID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return nil, huma.Error404NotFound("the requested book resource could not be found")
		default:
			return &ReturnBookTransactionOutput{}, err
		}
	}

	patron, err := app.models.Patrons.Get(sessionContext, data.PatronFilter{ID: &input.Body.PatronID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &ReturnBookTransactionOutput{}, huma.Error404NotFound("the requested patron resource could not be found")
		default:
			return &ReturnBookTransactionOutput{}, err
		}
	}

	transaction, err := app.models.Transactions.Get(sessionContext, data.TransactionFilter{
		Status:   ptr(data.TransactionStatusBorrowed),
		BookID:   &book.ID,
		PatronID: &patron.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &ReturnBookTransactionOutput{}, huma.Error404NotFound("the requested transaction resource could not be found")
		default:
			return &ReturnBookTransactionOutput{}, err
		}
	}

	transaction.ReturnedAt = time.Now()
	transaction.Status = data.TransactionStatusReturned

	if err = app.models.Transactions.Update(sessionContext, data.TransactionFilter{ID: &transaction.ID}, transaction); err != nil {
		return &ReturnBookTransactionOutput{}, err
	}

	book.BorrowedCopies = book.BorrowedCopies - input.Body.Copies
	err = app.models.Books.Update(sessionContext, data.BookFilter{ID: &book.ID}, book)
	if err != nil {
		return &ReturnBookTransactionOutput{}, err
	}

	if err = session.CommitTransaction(ctx); err != nil {
		return &ReturnBookTransactionOutput{}, err
	}

	var message string
	if input.Body.Copies > 1 {
		message = fmt.Sprintf("successfully returned %v copies of book with ISBN %v (id: %v)", input.Body.Copies, book.ISBN, book.ID)
	} else {
		message = fmt.Sprintf("successfully returned %v copy of book with ISBN %v (id: %v)", input.Body.Copies, book.ISBN, book.ID)
	}

	resp := &ReturnBookTransactionOutput{
		Body: message,
	}

	return resp, nil
}

// updateTransactionHandler handles a request to update an existing transaction by ID.
func (app *Application) updateTransactionHandler(ctx context.Context, input *UpdateTransactionInput) (*UpdateTransactionOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	transaction, err := app.models.Transactions.Get(ctx, data.TransactionFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return nil, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &UpdateTransactionOutput{}, err
		}
	}

	if input.Body.DueDate != nil {
		if transaction.Status == data.TransactionStatusReturned {
			return nil, huma.Error422UnprocessableEntity(fmt.Sprintf("Due date cannot be updated because the transaction status is %s", data.TransactionStatusReturned))
		}

		transaction.DueDate = *input.Body.DueDate
	}

	resp := &UpdateTransactionOutput{
		Body: *transaction,
	}

	return resp, nil
}

// deleteTransactionHandler handles a request to delete a transaction by ID.
func (app *Application) deleteTransactionHandler(ctx context.Context, input *DeleteTransactionInput) (*DeleteTransactionOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	err := app.models.Transactions.Delete(ctx, data.TransactionFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return nil, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &DeleteTransactionOutput{}, err
		}
	}

	resp := &DeleteTransactionOutput{
		Body: "transaction successfully deleted",
	}

	return resp, nil
}
