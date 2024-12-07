package main

import (
	"flag"
	"github.com/mzeevi/library/internal/generate"
	"log/slog"
	"os"
	"time"

	"github.com/mzeevi/library/internal/data"
)

type Library struct {
	Books   []data.Book
	Patrons []data.Patron
}

type config struct {
	cost struct {
		admission float64
		fine      float64
	}
	discount struct {
		teacher float64
		student float64
	}
	transactions struct {
		file   string
		format string
	}
}

type application struct {
	discounts    map[data.PatronCategoryType]float64
	transactions data.TransactionsOutput
}

func main() {
	var cfg config
	var app application

	flag.Float64Var(&cfg.cost.admission, "admission-price", 100, "Price for becoming library patron")
	flag.Float64Var(&cfg.cost.fine, "overdue-fine", 10, "Fine for returning overdue book")
	flag.Float64Var(&cfg.discount.teacher, "teacher-discount-percentage", 20, "Discount percentage for teachers")
	flag.Float64Var(&cfg.discount.student, "student-discount-discountPercentage", 25, "Discount percentage for students")
	flag.StringVar(&cfg.transactions.file, "transactions-output-file", "transactions-output", "Filename for the transactions output")
	flag.StringVar(&cfg.transactions.format, "transactions-output-format", "csv", "Format for the transactions output")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app.setDiscounts(cfg)
	app.setTransactionsOutput(cfg)

	err := app.transactions.CreateWriter(cfg.transactions.file, cfg.transactions.format)
	if err != nil {
		logger.Error("failed to create transaction writer", "error", err)
		os.Exit(1)
	}

	defer func(transactions data.TransactionsOutput) {
		err = transactions.CloseWriter()
		if err != nil {
			logger.Error("failed to close writer", "error", err)
			os.Exit(1)
		}
	}(app.transactions)

	transactionHeaders := []string{"Timestamp", "Type", "Patron Name", "Book Name"}
	err = app.transactions.WriteRecord(transactionHeaders)
	if err != nil {
		logger.Error("failed to write headers", "error", err)
		os.Exit(1)
	}

	books := generate.Books(10)
	patrons, err := generate.Patrons(10, app.discounts)
	if err != nil {
		logger.Error("failed generate patrons", "errors", err)
		os.Exit(1)
	}

	library := Library{
		Books:   books,
		Patrons: patrons,
	}

	err = library.Patrons[0].BorrowBook("test-book-1", books)
	if err != nil {
		logger.Error("failed to borrow book", "error", err)
		os.Exit(1)
	}

	err = app.transactions.WriteRecord(createTransactionRecord("borrow", library.Patrons[0].Name, "test-book-1"))
	if err != nil {
		logger.Error("failed to record transaction", "error", err)
		os.Exit(1)
	}

	err = library.Patrons[0].ReturnBook("test-book-1", books)
	if err != nil {
		logger.Error("failed to return book", "error", err)
		os.Exit(1)
	}

	err = app.transactions.WriteRecord(createTransactionRecord("return", library.Patrons[0].Name, "test-book-1"))
	if err != nil {
		logger.Error("failed to record transaction", "error", err)
		os.Exit(1)
	}

	searchTitle := "test-book"
	searchPublisher := []string{"test-publisher-2-4"}
	maxPages := 4
	minPages := 4
	searchBooks := data.SearchBooks(books, data.SearchCriteria{
		Title:      &searchTitle,
		MaxPages:   &maxPages,
		MinPages:   &minPages,
		Publishers: &searchPublisher,
	})

	logger.Info("All returned books", "searchedBooks", searchBooks)
}

// setDiscounts populates the discount fields inside the app struct.
func (app *application) setDiscounts(cfg config) {
	if cfg.discount.student < 0 || cfg.discount.student > 100 {
		slog.Error("student discount percentage must be between 0 and 100")
		os.Exit(1)
	}

	if cfg.discount.teacher < 0 || cfg.discount.teacher > 100 {
		slog.Error("teacher discount percentage must be between 0 and 100")
		os.Exit(1)
	}

	app.discounts = map[data.PatronCategoryType]float64{
		data.Student: cfg.discount.student,
		data.Teacher: cfg.discount.teacher,
	}
}

// setDiscounts populates the transaction fields inside the app struct.
func (app *application) setTransactionsOutput(cfg config) {
	switch data.TransactionOutputType(cfg.transactions.format) {
	case data.CSVOutputFormat:
		app.transactions = &data.CSVTransactionOutput{}
	case data.XLSMOutputFormat:
		app.transactions = &data.ExcelTransactionOutput{}
	case data.EXLAMOutputFormat:
		app.transactions = &data.ExcelTransactionOutput{}
	case data.XLSXOutputFormat:
		app.transactions = &data.ExcelTransactionOutput{}
	case data.XLTMOutputFormat:
		app.transactions = &data.ExcelTransactionOutput{}
	case data.XLTXOutputFormat:
		app.transactions = &data.ExcelTransactionOutput{}
	default:
		slog.Error("unsupported output format")
		os.Exit(1)
	}
}

// createTransactionRecord returns a slice containing information to write to an output file.
func createTransactionRecord(transactionType, patronName, bookName string) []string {
	return []string{
		time.Now().String(),
		transactionType,
		patronName,
		bookName,
	}
}
