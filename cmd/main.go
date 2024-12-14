package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/mzeevi/library/internal/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"os"
)

type config struct {
	cost struct {
		overdueFine float64
		discount    struct {
			teacher float64
			student float64
		}
	}
	output struct {
		enabled bool
		file    string
		format  string
	}
	db struct {
		dsn                    string
		database               string
		booksCollection        string
		patronsCollection      string
		transactionsCollection string
	}
}

type application struct {
	models data.Models
	cost   struct {
		overdueFine float64
		discounts   map[data.PatronCategoryType]float64
	}
	transactions data.Output
	logger       *slog.Logger
}

func main() {
	var cfg config
	var app application

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "MongoDB DSN")
	flag.StringVar(&cfg.db.database, "db", "library", "MongoDB Database name")
	flag.StringVar(&cfg.db.booksCollection, "books-collection", "books", "MongoDB collection name for books")
	flag.StringVar(&cfg.db.patronsCollection, "patrons-collection", "patrons", "MongoDB collection name for patrons")
	flag.StringVar(&cfg.db.transactionsCollection, "transactions-collection", "transactions", "MongoDB collection name for output")

	flag.Float64Var(&cfg.cost.overdueFine, "overdue-fine", 10, "Fine for returning overdue book")
	flag.Float64Var(&cfg.cost.discount.teacher, "teacher-discount-percentage", 20, "Discount percentage for teachers")
	flag.Float64Var(&cfg.cost.discount.student, "student-discount-discountPercentage", 25, "Discount percentage for students")

	flag.BoolVar(&cfg.output.enabled, "output-enabled", false, "Flag to enable writing to output file")
	flag.StringVar(&cfg.output.file, "output-file", "output", "Filename for the output file")
	flag.StringVar(&cfg.output.format, "output-format", "csv", "Format for the output file")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbClient, err := initDBClient(cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initiate db dbClient: %v", err))
		os.Exit(1)
	}
	defer func() {
		if err = dbClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err = app.set(cfg, dbClient, logger); err != nil {
		logger.Error(fmt.Sprintf("failed to set app values: %v", err))
		os.Exit(1)
	}

	if err = app.models.Books.CreateUniqueISBNIndex(); err != nil {
		logger.Error(fmt.Sprintf("failed to create unique index: %v", err))
		os.Exit(1)
	}

	if err = app.basic(); err != nil {
		logger.Error(fmt.Sprintf("failed to check basic functionality: %v", err))
		os.Exit(1)
	}
}

// initDBClient initializes a client to the database.
func initDBClient(cfg config) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(cfg.db.dsn).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// set populates the fields of the application struct.
func (app *application) set(cfg config, dbClient *mongo.Client, logger *slog.Logger) error {
	if cfg.output.enabled {
		app.setOutput(cfg)
	}

	if err := app.setCost(cfg); err != nil {
		return fmt.Errorf("failed to set discounts: %v", err)
	}

	app.models = data.NewModels(dbClient, cfg.db.database, map[string]string{
		data.BooksCollectionKey:        cfg.db.booksCollection,
		data.PatronsCollectionKey:      cfg.db.patronsCollection,
		data.TransactionsCollectionKey: cfg.db.transactionsCollection,
	})

	app.logger = logger

	return nil
}

// setCost populates the discount fields inside the app struct.
func (app *application) setCost(cfg config) error {
	if cfg.cost.discount.student < 0 || cfg.cost.discount.student > 100 {
		return fmt.Errorf("student discount percentage must be between 0 and 100")
	}

	if cfg.cost.discount.teacher < 0 || cfg.cost.discount.teacher > 100 {
		return fmt.Errorf("teacher discount percentage must be between 0 and 100")
	}

	app.cost.overdueFine = cfg.cost.overdueFine
	app.cost.discounts = map[data.PatronCategoryType]float64{
		data.Student: cfg.cost.discount.student,
		data.Teacher: cfg.cost.discount.teacher,
	}

	return nil
}

// setOutput populates the transaction fields inside the app struct.
func (app *application) setOutput(cfg config) {
	switch data.OutputType(cfg.output.format) {
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
