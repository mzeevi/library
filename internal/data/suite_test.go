package data

import (
	"context"
	"github.com/mzeevi/library/internal/data/testhelpers"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

func TestSuiteLibrary(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

type TestSuite struct {
	suite.Suite
	mdbContainer *testhelpers.MongoDBContainer
	models       *Models
	ctx          context.Context
}

// SetupSuite sets up the testing suite.
func (ts *TestSuite) SetupSuite() {
	ts.ctx = context.Background()

	mdbContainer, err := testhelpers.CreateMongoDBContainer(ts.ctx)
	if err != nil {
		log.Fatal(err)
	}

	ts.mdbContainer = mdbContainer

	client, err := mdbContainer.Client(ts.ctx)
	if err != nil {
		log.Fatal(err)
	}

	ts.models = &Models{
		Books:        BookModel{Client: client, Database: "test-library", Collection: BooksCollectionKey},
		Patrons:      PatronModel{Client: client, Database: "test-library", Collection: PatronsCollectionKey},
		Transactions: TransactionModel{Client: client, Database: "test-library", Collection: TransactionsCollectionKey},
	}

	if err = ts.populateBooksInDB(); err != nil {
		log.Fatal(err)
	}

	if err = ts.populatePatronsInDB(); err != nil {
		log.Fatal(err)
	}

	if err = ts.populateTransactionsInDB(); err != nil {
		log.Fatal(err)
	}
}

// TearDownSuite performs clean up.
func (ts *TestSuite) TearDownSuite() {
	if err := ts.mdbContainer.Terminate(ts.ctx); err != nil {
		log.Fatalf("error terminating mongodb container: %s", err)
	}
}

// ptr is a generic helper function for creating a pointer to any type.
func ptr[T any](v T) *T {
	return &v
}
