package api

import (
	"context"
	"fmt"
	"github.com/mzeevi/library/internal/data"
	"github.com/mzeevi/library/internal/data/testhelpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

// InsertBooks inserts Books into the database.
func (app *Application) InsertBooks(n int) ([]string, error) {
	var ids []string

	client := app.Models.Books.Client
	coll := client.Database(app.Models.Books.Database).Collection(app.Models.Books.Collection)

	books := mockBooks(n)

	res, err := coll.InsertMany(context.Background(), books)
	if err != nil {
		return ids, err
	}

	for _, oid := range res.InsertedIDs {
		ids = append(ids, oid.(primitive.ObjectID).Hex())
	}

	return ids, nil
}

// InsertPatrons inserts Patrons into the database.
func (app *Application) InsertPatrons(n int) ([]string, error) {
	var ids []string

	client := app.Models.Patrons.Client
	coll := client.Database(app.Models.Patrons.Database).Collection(app.Models.Patrons.Collection)

	patrons, err := mockPatrons(n)
	if err != nil {
		return ids, fmt.Errorf("failed to generate patrons: %v", err)
	}

	res, err := coll.InsertMany(context.Background(), patrons)
	if err != nil {
		return ids, err
	}

	for _, oid := range res.InsertedIDs {
		ids = append(ids, oid.(primitive.ObjectID).Hex())
	}

	return ids, nil
}

// mockBooks returns a slice of n new mock Books.
func mockBooks(n int) []interface{} {
	var books []interface{}

	for i := 1; i <= n; i++ {
		s := strconv.Itoa(i)

		book := data.NewBook("", fmt.Sprintf("test-book-%s", s), testhelpers.GenerateISBN(),
			i, i, i,
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
func mockPatrons(n int) ([]interface{}, error) {
	var patrons []interface{}
	var category string

	for i := 1; i <= n; i++ {
		s := strconv.Itoa(i)

		switch i % 2 {
		case 0:
			category = "teacher"
		default:
			category = "student"
		}

		patron := data.NewPatron("", fmt.Sprintf("test-%s", s), fmt.Sprintf("test-%s@email.com", s), category)
		patrons = append(patrons, patron)
	}

	return patrons, nil
}
