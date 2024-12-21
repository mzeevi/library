package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-chi/httplog/v2"
	"github.com/mzeevi/library/internal/api"
	"github.com/mzeevi/library/internal/database"
	"log/slog"
	"os"
	"strings"
	"time"
)

func main() {
	var app api.Application

	flag.IntVar(&app.Config.Port, "port", 8080, "API server port")

	flag.StringVar(&app.Config.DB.DSN, "db-dsn", "", "MongoDB DSN")
	flag.StringVar(&app.Config.DB.Database, "db", "library", "MongoDB Database name")
	flag.StringVar(&app.Config.DB.BooksCollection, "books-collection", "books", "MongoDB collection name for books")
	flag.StringVar(&app.Config.DB.PatronsCollection, "patrons-collection", "patrons", "MongoDB collection name for patrons")
	flag.StringVar(&app.Config.DB.TransactionsCollection, "transactions-collection", "transactions", "MongoDB collection name for transactions")
	flag.StringVar(&app.Config.DB.TokensCollection, "tokens-collection", "tokens", "MongoDB collection name for tokens")
	flag.StringVar(&app.Config.DB.AdminsCollection, "admins-collection", "admins", "MongoDB collection name for admins")

	flag.BoolVar(&app.Config.Admin.Create, "create-admin", true, "create admin user")
	flag.StringVar(&app.Config.Admin.Username, "admin-username", "", "admin user")
	flag.StringVar(&app.Config.Admin.Password, "admin-password", "", "admin password")

	flag.Float64Var(&app.Config.Cost.OverdueFine, "overdue-fine", 10, "Fine for returning overdue book")
	flag.Float64Var(&app.Config.Cost.Discount.Teacher, "teacher-discount-percentage", 20, "Discount percentage for teachers")
	flag.Float64Var(&app.Config.Cost.Discount.Student, "student-discount-discountPercentage", 25, "Discount percentage for students")

	flag.BoolVar(&app.Config.Output.Enabled, "output-enabled", false, "Flag to enable writing to output file")
	flag.StringVar(&app.Config.Output.File, "output-file", "output", "Filename for the output file")
	flag.StringVar(&app.Config.Output.Format, "output-format", "csv", "Format for the output file")

	flag.StringVar(&app.Config.JTW.Secret, "jwt-secret", "", "JWT secret")
	flag.StringVar(&app.Config.JTW.Issuer, "jwt-issuer", "library.com", "JWT secret")
	flag.StringVar(&app.Config.JTW.Audience, "jwt-audience", "library.com", "JWT secret")

	flag.BoolVar(&app.Config.Demo.Patrons, "demo-patrons", false, "create demo patrons")
	flag.BoolVar(&app.Config.Demo.Patrons, "demo-books", true, "create demo books")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		app.Config.CORS.TrustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	logger := httplog.NewLogger(app.Config.DB.Database, httplog.Options{
		JSON:             false,
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
		QuietDownPeriod:  10 * time.Second,
		SourceFieldName:  "source",
	})

	dbClient, err := database.Client(app.Config.DB.DSN)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initiate database client: %v", err))
		os.Exit(1)
	}
	defer func() {
		if err = dbClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err = app.Setup(dbClient, logger); err != nil {
		logger.Error(fmt.Sprintf("failed to set app values: %v", err))
		os.Exit(1)
	}

	if app.Config.Admin.Create {
		if err = app.Models.Admins.New(context.Background(), app.Config.Admin.Username, app.Config.Admin.Password); err != nil {
			logger.Error(fmt.Sprintf("failed to create admin: %v", err))
			os.Exit(1)
		}
	}

	if app.Config.Demo.Patrons {
		_, err = app.InsertPatrons(10)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to insert demo patrons to database: %v", err))
			os.Exit(1)
		}
	}

	if app.Config.Demo.Books {
		_, err = app.InsertBooks(10)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to insert demo books to database: %v", err))
			os.Exit(1)
		}
	}

	if err = app.Serve(); err != nil {
		logger.Error(fmt.Sprintf("failed to set up router: %v", err))
		os.Exit(1)
	}
}
