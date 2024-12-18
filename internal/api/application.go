package api

import (
	"fmt"
	"github.com/go-chi/httplog/v2"
	"github.com/mzeevi/library/internal/config"
	"github.com/mzeevi/library/internal/data"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	Config config.Input
	models data.Models
	cost   struct {
		overdueFine float64
		discounts   map[string]float64
	}
	transactions data.Output
	logger       *httplog.Logger
}

// Setup populates the fields of the Application struct.
func (app *Application) Setup(dbClient *mongo.Client, logger *httplog.Logger) error {
	app.logger = logger
	cfg := app.Config

	if cfg.Output.Enabled {
		if err := app.setupOutput(app.Config.Output.Format); err != nil {
			return fmt.Errorf("failed to setup output: %v", err)
		}
	}

	if err := app.setupCost(cfg.Cost.Discount.Student, cfg.Cost.Discount.Teacher, cfg.Cost.OverdueFine); err != nil {
		return fmt.Errorf("failed to setup discounts: %v", err)
	}

	if err := app.setupModels(dbClient, cfg.DB.Database, cfg.DB.BooksCollection, cfg.DB.PatronsCollection, cfg.DB.TransactionsCollection); err != nil {
		return fmt.Errorf("failed to setup models: %v", err)
	}

	return nil
}

// setupCost populates the discount fields inside the app struct.
func (app *Application) setupCost(studentDiscountPercent, teacherDiscountPercent, overdueFine float64) error {
	if studentDiscountPercent < 0 || studentDiscountPercent > 100 {
		return fmt.Errorf("student discount percentage must be between 0 and 100")
	}

	if teacherDiscountPercent < 0 || teacherDiscountPercent > 100 {
		return fmt.Errorf("teacher discount percentage must be between 0 and 100")
	}

	app.cost.overdueFine = overdueFine
	app.cost.discounts = map[string]float64{
		studentCategory: studentDiscountPercent,
		teacherCategory: teacherDiscountPercent,
	}

	return nil
}

// setupOutput populates the transaction fields inside the app struct.
func (app *Application) setupOutput(format string) error {
	switch data.OutputType(format) {
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
		return fmt.Errorf("unsupported output format")
	}

	return nil
}

// setupModels populates the model fields inside the app struct.
func (app *Application) setupModels(dbClient *mongo.Client, dbName, booksCollection, patronsCollection, transactionCollection string) error {
	app.models = data.NewModels(dbClient, dbName, map[string]string{
		data.BooksCollectionKey:        booksCollection,
		data.PatronsCollectionKey:      patronsCollection,
		data.TransactionsCollectionKey: transactionCollection,
	})

	if err := app.models.Books.CreateUniqueIndex(); err != nil {
		return fmt.Errorf("failed to create unique index: %v", err)
	}

	if err := app.models.Patrons.CreateUniqueIndex(); err != nil {
		return fmt.Errorf("failed to create unique index: %v", err)
	}

	return nil
}
