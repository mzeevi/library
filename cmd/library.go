package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mzeevi/library/internal/data"
	"github.com/mzeevi/library/internal/data/testhelpers"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"time"
)

func (app *application) basic() error {
	if err := app.insertBooks(); err != nil {
		return fmt.Errorf("failed to insert books to database: %v", err)
	}

	if err := app.insertPatrons(); err != nil {
		return fmt.Errorf("failed to insert patrons to database: %v", err)
	}

	if err := app.borrowBook("1", "1", time.Now()); err != nil {
		return fmt.Errorf("failed to borrow book: %v", err)
	}

	if err := app.borrowBook("2", "1", time.Now()); err != nil {
		return fmt.Errorf("failed to borrow book: %v", err)
	}

	if err := app.borrowBook("3", "3", time.Now()); err != nil {
		return fmt.Errorf("failed to borrow book: %v", err)
	}

	if err := app.returnBook("1", "1"); err != nil {
		return fmt.Errorf("failed to return book: %v", err)
	}

	if err := app.borrowBook("1", "1", time.Now()); err != nil {
		return fmt.Errorf("failed to borrow book: %v", err)
	}

	if err := app.getAllBorrowedBooks(); err != nil {
		return fmt.Errorf("failed to get all borrowed books: %v", err)
	}

	if err := app.getAllPatrons(); err != nil {
		return fmt.Errorf("failed to get all patrons: %v", err)
	}

	if err := app.getAllTransactions(); err != nil {
		return fmt.Errorf("failed to get all output: %v", err)
	}

	if err := app.calculateFine("3"); err != nil {
		return fmt.Errorf("failed to get all output: %v", err)
	}

	return nil
}

// insertBooks inserts Books into the database.
func (app *application) insertBooks() error {
	client := app.models.Books.Client

	coll := client.Database(app.models.Books.Database).Collection(app.models.Books.Collection)
	books := mockBooks(10)

	_, err := coll.InsertMany(context.Background(), books)
	if err != nil {
		return err
	}

	return nil
}

// insertPatrons inserts Patrons into the database.
func (app *application) insertPatrons() error {
	client := app.models.Patrons.Client

	coll := client.Database(app.models.Patrons.Database).Collection(app.models.Patrons.Collection)
	patrons, err := mockPatrons(10, app.cost.discounts)
	if err != nil {
		return fmt.Errorf("failed to generate patrons: %v", err)
	}

	_, err = coll.InsertMany(context.Background(), patrons)
	if err != nil {
		return err
	}

	return nil
}

// borrowBook implements the logic of borrowing a book.
func (app *application) borrowBook(bookID, patronID string, dueDate time.Time) error {
	dbClient := app.models.Transactions.Client

	session, err := dbClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	sessionContext := mongo.NewSessionContext(context.Background(), session)

	if err = session.StartTransaction(); err != nil {
		return err
	}

	_, err = app.models.Books.Get(sessionContext, data.BookFilter{ID: &bookID})
	if err != nil {
		return err
	}

	_, err = app.models.Patrons.Get(sessionContext, data.PatronFilter{ID: &patronID})
	if err != nil {
		return err
	}

	var available bool

	_, err = app.models.Transactions.Get(sessionContext, data.TransactionFilter{
		Status: ptr(data.TransactionStatusBorrowed),
		BookID: &bookID,
	})
	if err != nil {
		if errors.Is(err, data.ErrDocumentNotFound) {
			available = true
		}
	}

	if !available {
		return errors.New("book already borrowed")
	}

	transaction := &data.Transaction{
		PatronID:   patronID,
		BookID:     bookID,
		BorrowedAt: time.Now(),
		DueDate:    dueDate,
		Status:     data.TransactionStatusBorrowed,
	}

	if _, err = app.models.Transactions.Insert(sessionContext, transaction); err != nil {
		return err
	}

	return session.CommitTransaction(context.Background())
}

// returnBook implements the logic of returning a book.
func (app *application) returnBook(bookID, patronID string) error {
	dbClient := app.models.Transactions.Client

	session, err := dbClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	sessionContext := mongo.NewSessionContext(context.Background(), session)

	if err = session.StartTransaction(); err != nil {
		return err
	}

	_, err = app.models.Books.Get(sessionContext, data.BookFilter{ID: &bookID})
	if err != nil {
		return err
	}

	_, err = app.models.Patrons.Get(sessionContext, data.PatronFilter{ID: &patronID})
	if err != nil {
		return err
	}

	transaction, err := app.models.Transactions.Get(sessionContext, data.TransactionFilter{
		Status: ptr(data.TransactionStatusBorrowed),
		BookID: &bookID,
	})
	if err != nil {
		return err
	}

	transaction.Status = data.TransactionStatusReturned

	if err = app.models.Transactions.Update(sessionContext, data.TransactionFilter{ID: &transaction.ID}, transaction); err != nil {
		return err
	}

	return session.CommitTransaction(context.Background())
}

// getAllTransactions prints all the borrowed books from the database.
func (app *application) getAllBorrowedBooks() error {
	transactions, _, err := app.models.Transactions.GetAll(
		context.Background(),
		data.TransactionFilter{
			Status: ptr(data.TransactionStatusBorrowed),
		},
		data.Paginator{})
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		app.logger.Info("got book", "bookID", transaction.BookID, "patronID", transaction.PatronID)
	}

	return nil
}

// getAllTransactions prints all the patrons from the database.
func (app *application) getAllPatrons() error {
	paginatior := data.Paginator{
		PageSize: 5,
		Page:     4,
	}

	patrons, _, err := app.models.Patrons.GetAll(context.Background(), data.PatronFilter{}, paginatior)
	if err != nil {
		return err
	}

	for _, patron := range patrons {
		app.logger.Info("got patron", "patronID", patron.ID, "patronName", patron.Name)
	}

	return nil
}

// getAllTransactions prints all the transactions from the database.
func (app *application) getAllTransactions() error {
	paginator := data.Paginator{
		PageSize: 5,
		Page:     1,
	}

	transactions, _, err := app.models.Transactions.GetAll(context.Background(), data.TransactionFilter{}, paginator)
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		app.logger.Info("got transaction", "transactionID", transaction.ID)
	}

	return nil
}

// calculateFine prints the fine of a patron on overdue books.
func (app *application) calculateFine(patronID string) error {
	var totalFine float64

	patron, err := app.models.Patrons.Get(context.Background(), data.PatronFilter{ID: &patronID})
	if err != nil {
		return err
	}

	transactions, _, err := app.models.Transactions.GetAll(context.Background(), data.TransactionFilter{PatronID: &patronID}, data.Paginator{})
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		daysOverdue := time.Since(transaction.DueDate).Hours() / 24
		if daysOverdue > 0 {
			totalFine += daysOverdue * app.cost.overdueFine
		}

		app.logger.Info(fmt.Sprintf("got transaction - %s | borrowed at - %v | due date - %v", transaction.ID, transaction.BorrowedAt, transaction.DueDate))
	}

	app.logger.Info("calculated fine", "patronID", patronID, "category", patron.Category.Type(), "fine", totalFine)

	return nil
}

// mockBooks returns a slice of n new mock Books.
func mockBooks(n int) []interface{} {
	var books []interface{}

	for i := 1; i <= n; i++ {
		s := strconv.Itoa(i)

		book := data.NewBook(strconv.Itoa(i), fmt.Sprintf("test-book-%s", s), testhelpers.GenerateISBN(),
			i, i,
			[]string{fmt.Sprintf("test-author-1-%s", s), fmt.Sprintf("test-author-2-%s", s)},
			[]string{fmt.Sprintf("test-publisher-1-%s", s), fmt.Sprintf("test-publisher-2-%s", s)},
			[]string{fmt.Sprintf("test-genre-1-%s", s), fmt.Sprintf("test-genre-2-%s", s)},
			time.Date(2015, time.November, 20, 0, 0, 0, 0, time.UTC),
		)

		books = append(books, book)
	}

	return books
}

// mockPatrons returns a slice of n new mock Patrons.
func mockPatrons(n int, discounts map[data.PatronCategoryType]float64) ([]interface{}, error) {
	var patrons []interface{}
	var category data.PatronCategory

	for i := 1; i <= n; i++ {
		s := strconv.Itoa(i)

		switch i % 2 {
		case 0:
			category = data.TeacherCategory{
				CategoryType:       data.Teacher,
				DiscountPercentage: discounts[data.Teacher],
			}
		default:
			category = data.StudentCategory{
				CategoryType:       data.Student,
				DiscountPercentage: discounts[data.Student],
			}
		}

		patron := data.NewPatron(strconv.Itoa(i), fmt.Sprintf("test-%s", s), fmt.Sprintf("test-%s@email.com", s), category)
		patrons = append(patrons, patron)
	}

	return patrons, nil
}

// ptr is a generic helper function for creating a pointer to any type.
func ptr[T any](v T) *T {
	return &v
}
