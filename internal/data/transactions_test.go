package data

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

var (
	testTransactionsIDs []interface{}
	testTransactions    = []interface{}{
		NewTransaction("", "1", "B1", TransactionStatusBorrowed, time.Date(2024, time.December, 10, 14, 0, 0, 0, time.UTC), time.Date(2024, time.December, 12, 14, 0, 0, 0, time.UTC)),
		NewTransaction("", "2", "B2", TransactionStatusReturned, time.Date(2024, time.December, 9, 9, 0, 0, 0, time.UTC), time.Date(2024, time.December, 11, 9, 0, 0, 0, time.UTC)),
		NewTransaction("", "3", "B3", TransactionStatusBorrowed, time.Date(2024, time.December, 8, 16, 0, 0, 0, time.UTC), time.Date(2024, time.December, 10, 16, 0, 0, 0, time.UTC)),
		NewTransaction("", "4", "B4", TransactionStatusReturned, time.Date(2024, time.December, 7, 10, 30, 0, 0, time.UTC), time.Date(2024, time.December, 9, 10, 30, 0, 0, time.UTC)),
		NewTransaction("", "5", "B5", TransactionStatusBorrowed, time.Date(2024, time.December, 6, 11, 45, 0, 0, time.UTC), time.Date(2024, time.December, 8, 11, 45, 0, 0, time.UTC)),
		NewTransaction("", "6", "B6", TransactionStatusReturned, time.Date(2024, time.December, 5, 15, 0, 0, 0, time.UTC), time.Date(2024, time.December, 7, 15, 0, 0, 0, time.UTC)),
		NewTransaction("", "7", "B7", TransactionStatusBorrowed, time.Date(2024, time.December, 4, 13, 0, 0, 0, time.UTC), time.Date(2024, time.December, 6, 13, 0, 0, 0, time.UTC)),
		NewTransaction("", "8", "B8", TransactionStatusBorrowed, time.Date(2024, time.December, 3, 12, 30, 0, 0, time.UTC), time.Date(2024, time.December, 5, 12, 30, 0, 0, time.UTC)),
		NewTransaction("", "9", "B9", TransactionStatusReturned, time.Date(2024, time.December, 2, 8, 0, 0, 0, time.UTC), time.Date(2024, time.December, 4, 8, 0, 0, 0, time.UTC)),
		NewTransaction("", "10", "B10", TransactionStatusBorrowed, time.Date(2024, time.December, 1, 17, 0, 0, 0, time.UTC), time.Date(2024, time.December, 3, 17, 0, 0, 0, time.UTC)),
		NewTransaction("conflict", "11", "B11", TransactionStatusBorrowed, time.Date(2024, time.December, 10, 14, 0, 0, 0, time.UTC), time.Date(2024, time.December, 12, 14, 0, 0, 0, time.UTC)),
	}
)

// populateTransactionsInDB inserts test Transactions into the DB.
func (ts *TestSuite) populateTransactionsInDB() error {
	client := ts.models.Transactions.Client
	coll := client.Database(ts.models.Transactions.Database).Collection(ts.models.Transactions.Collection)

	res, err := coll.InsertMany(ts.ctx, testTransactions)
	if err != nil {
		return err
	}

	testTransactionsIDs = res.InsertedIDs

	return nil
}

// deleteTransactionsFromDB deletes test Transactions from the DB.
func (ts *TestSuite) deleteTransactionsFromDB(filter TransactionFilter) error {
	client := ts.models.Transactions.Client
	coll := client.Database(ts.models.Transactions.Database).Collection(ts.models.Transactions.Collection)

	queryFilter, err := buildTransactionFilter(filter)
	if err != nil {
		return err
	}

	_, err = coll.DeleteMany(ts.ctx, queryFilter)

	return err
}

func (ts *TestSuite) TestTransactionInsert() {
	t := ts.T()

	tests := []struct {
		name        string
		transaction Transaction
		expectError bool
	}{
		{
			name: "ValidTransaction",
			transaction: Transaction{
				PatronID:   "100000",
				BookID:     "10000B",
				BorrowedAt: time.Now(),
				DueDate:    time.Now(),
				Status:     TransactionStatusBorrowed,
			},
			expectError: false,
		},
		{
			name: "TransactionWithConflictingID",
			transaction: Transaction{
				ID:       "conflict",
				PatronID: "1",
				BookID:   "1B",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			id, err := ts.models.Transactions.Insert(ts.ctx, &tt.transaction)
			if tt.expectError {
				assert.ErrorIs(t, err, ErrDuplicateID)
			} else {
				assert.NoError(t, err)
				err = ts.deleteTransactionsFromDB(TransactionFilter{ID: &id})
				assert.NoError(t, err)
			}
		})
	}
}

func (ts *TestSuite) TestTransactionGet() {
	t := ts.T()

	tests := []struct {
		name        string
		filter      TransactionFilter
		expectError bool
		expectedID  string
	}{
		{
			name:        "ValidTransactionByID",
			filter:      TransactionFilter{ID: ptr(testTransactionsIDs[0].(primitive.ObjectID).Hex())},
			expectError: false,
			expectedID:  testTransactionsIDs[0].(primitive.ObjectID).Hex(),
		},
		{
			name:        "ValidTransactionByPatronID",
			filter:      TransactionFilter{PatronID: ptr("1")},
			expectError: false,
			expectedID:  testTransactionsIDs[0].(primitive.ObjectID).Hex(),
		},
		{
			name:        "ValidTransactionByBookID",
			filter:      TransactionFilter{BookID: ptr("B2")},
			expectError: false,
			expectedID:  testTransactionsIDs[1].(primitive.ObjectID).Hex(),
		},
		{
			name: "TransactionNotFound",
			filter: TransactionFilter{
				ID: ptr("nonexistentID"),
			},
			expectError: true,
			expectedID:  "",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			transaction, err := ts.models.Transactions.Get(ts.ctx, tt.filter)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, transaction)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transaction)
				assert.Equal(t, tt.expectedID, transaction.ID)
			}
		})
	}
}

func (ts *TestSuite) TestGetAllTransactions() {
	t := ts.T()

	tests := []struct {
		name             string
		filter           TransactionFilter
		paginator        Paginator
		sorter           Sorter
		expectedIDs      []string
		expectedCount    int
		expectedMetadata Metadata
		expectError      bool
	}{
		{
			name: "FilterByPatronID",
			filter: TransactionFilter{
				PatronID: ptr("1"),
			},
			paginator:     Paginator{Page: 1, PageSize: 5},
			expectedIDs:   []string{testTransactionsIDs[0].(primitive.ObjectID).Hex()},
			expectedCount: 1,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     5,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectError: false,
		},
		{
			name: "FilterByBookID",
			filter: TransactionFilter{
				BookID: ptr("B2"),
			},
			paginator:     Paginator{Page: 1, PageSize: 3},
			expectedIDs:   []string{testTransactionsIDs[1].(primitive.ObjectID).Hex()},
			expectedCount: 1,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     3,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectError: false,
		},
		{
			name: "FilterByMinBorrowedAt",
			filter: TransactionFilter{
				MinBorrowedAt: ptr(time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator: Paginator{Page: 1, PageSize: 2},
			expectedIDs: []string{
				testTransactionsIDs[0].(primitive.ObjectID).Hex(),
				testTransactionsIDs[1].(primitive.ObjectID).Hex(),
			},
			expectedCount: 2,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     6,
				TotalRecords: 11,
			},
			expectError: false,
		},
		{
			name: "FilterByMaxDueDate",
			filter: TransactionFilter{
				MaxDueDate: ptr(time.Date(2023, time.April, 30, 0, 0, 0, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 2},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByStatus",
			filter: TransactionFilter{
				Status: ptr(TransactionStatusReturned),
			},
			paginator: Paginator{Page: 1, PageSize: 4},
			expectedIDs: []string{
				testTransactionsIDs[1].(primitive.ObjectID).Hex(),
				testTransactionsIDs[3].(primitive.ObjectID).Hex(),
				testTransactionsIDs[5].(primitive.ObjectID).Hex(),
				testTransactionsIDs[8].(primitive.ObjectID).Hex(),
			},
			expectedCount: 4,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     4,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 4,
			},
			expectError: false,
		},
		{
			name: "FilterByCreatedAtRange",
			filter: TransactionFilter{
				MinCreatedAt: ptr(time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)),
				MaxCreatedAt: ptr(time.Date(2023, time.March, 31, 23, 59, 59, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByUpdatedAtRange",
			filter: TransactionFilter{
				MinUpdatedAt: ptr(time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC)),
				MaxUpdatedAt: ptr(time.Date(2023, time.September, 30, 23, 59, 59, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 5},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByReturnedAtRange",
			filter: TransactionFilter{
				MinReturnedAt: ptr(time.Date(2023, time.November, 10, 0, 0, 0, 0, time.UTC)),
				MaxReturnedAt: ptr(time.Date(2023, time.December, 10, 0, 0, 0, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 5},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByExactVersion",
			filter: TransactionFilter{
				Version: ptr(int32(0)),
			},
			paginator: Paginator{Page: 2, PageSize: 5},
			expectedIDs: []string{
				testTransactionsIDs[5].(primitive.ObjectID).Hex(),
				testTransactionsIDs[6].(primitive.ObjectID).Hex(),
				testTransactionsIDs[7].(primitive.ObjectID).Hex(),
				testTransactionsIDs[8].(primitive.ObjectID).Hex(),
				testTransactionsIDs[9].(primitive.ObjectID).Hex(),
			},
			expectedCount: 5,
			expectedMetadata: Metadata{
				CurrentPage:  2,
				PageSize:     5,
				FirstPage:    1,
				LastPage:     3,
				TotalRecords: 11,
			},
			expectError: false,
		},
		{
			name: "FilterByID",
			filter: TransactionFilter{
				ID: ptr(testTransactionsIDs[5].(primitive.ObjectID).Hex()),
			},
			paginator:     Paginator{Page: 1, PageSize: 5},
			expectedIDs:   []string{testTransactionsIDs[5].(primitive.ObjectID).Hex()},
			expectedCount: 1,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     5,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectError: false,
		},
		{
			name: "FilterByMultipleParams",
			filter: TransactionFilter{
				PatronID:   ptr(testTransactionsIDs[1].(primitive.ObjectID).Hex()),
				MinDueDate: ptr(time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)),
				MaxDueDate: ptr(time.Date(2023, time.March, 31, 23, 59, 59, 0, time.UTC)),
				Status:     ptr(TransactionStatusReturned),
			},
			paginator:        Paginator{Page: 1, PageSize: 3},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByBorrowedAtRange",
			filter: TransactionFilter{
				MinBorrowedAt: ptr(time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)),
				MaxBorrowedAt: ptr(time.Date(2023, time.March, 31, 23, 59, 59, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 4},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name:      "FilterWithEmptyFields",
			filter:    TransactionFilter{},
			paginator: Paginator{Page: 1, PageSize: 5},
			expectedIDs: []string{
				testTransactionsIDs[0].(primitive.ObjectID).Hex(),
				testTransactionsIDs[1].(primitive.ObjectID).Hex(),
				testTransactionsIDs[2].(primitive.ObjectID).Hex(),
				testTransactionsIDs[3].(primitive.ObjectID).Hex(),
				testTransactionsIDs[4].(primitive.ObjectID).Hex(),
			},
			expectedCount: 5,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     5,
				FirstPage:    1,
				LastPage:     3,
				TotalRecords: 11,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactions, metadata, err := ts.models.Transactions.GetAll(ts.ctx, tt.filter, tt.paginator, tt.sorter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, transactions, tt.expectedCount)

				var actualIDs []string
				for _, transaction := range transactions {
					actualIDs = append(actualIDs, transaction.ID)
				}
				assert.ElementsMatch(t, tt.expectedIDs, actualIDs)

				assert.Equal(t, tt.expectedMetadata, metadata)
			}
		})
	}
}

func (ts *TestSuite) TestUpdateTransaction() {
	t := ts.T()

	tests := []struct {
		name               string
		initialTransaction Transaction
		updateData         Transaction
		expectError        bool
		expected           *Transaction
	}{
		{
			name: "UpdateExistingTransaction",
			initialTransaction: Transaction{
				PatronID:   "1000",
				BookID:     "B1000",
				BorrowedAt: time.Now(),
				DueDate:    time.Now(),
				Status:     TransactionStatusBorrowed,
			},
			updateData: Transaction{
				PatronID:   "1000",
				BookID:     "B1000",
				ReturnedAt: time.Now(),
				Status:     TransactionStatusReturned,
			},
			expectError: false,
			expected: &Transaction{
				PatronID:   "1000",
				BookID:     "B1000",
				ReturnedAt: time.Now(),
				Status:     TransactionStatusReturned,
			},
		},
		{
			name: "NonExistingBook",
			initialTransaction: Transaction{
				ID: "995cb5a4d3ddbde5ebeecc1f",
			},
			updateData: Transaction{
				PatronID: "Non-existent",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			if tt.expectError {
				err := ts.models.Transactions.Update(ts.ctx, TransactionFilter{ID: &tt.initialTransaction.ID}, &tt.updateData)
				assert.ErrorIs(t, err, ErrEditConflict)
			} else {
				id, err := ts.models.Transactions.Insert(ts.ctx, &tt.initialTransaction)
				assert.NoError(t, err)

				err = ts.models.Transactions.Update(ts.ctx, TransactionFilter{ID: &id}, &tt.updateData)
				assert.NoError(t, err)

				updatedTransaction, err := ts.models.Transactions.Get(ts.ctx, TransactionFilter{ID: &id})
				assert.NoError(t, err)

				assert.Equal(t, updatedTransaction.Version, int32(1))

				err = ts.deleteTransactionsFromDB(TransactionFilter{ID: &id})
				assert.NoError(t, err)
			}
		})
	}
}

func (ts *TestSuite) TestDeleteTransaction() {
	t := ts.T()

	tests := []struct {
		name               string
		initialTransaction Transaction
		expectError        bool
	}{
		{
			name: "ExistingTransaction",
			initialTransaction: Transaction{
				BookID:   "B19",
				PatronID: "19",
			},
			expectError: false,
		},
		{
			name: "NonExistingTransaction",
			initialTransaction: Transaction{
				ID: "995cb5a4d3ddbde5ebeecc1f",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			if tt.expectError {
				err := ts.models.Transactions.Delete(ts.ctx, TransactionFilter{ID: &tt.initialTransaction.ID})
				assert.ErrorIs(t, err, ErrDocumentNotFound)
			} else {
				id, err := ts.models.Transactions.Insert(ts.ctx, &tt.initialTransaction)
				assert.NoError(t, err)

				err = ts.models.Transactions.Delete(ts.ctx, TransactionFilter{ID: &id})
				assert.NoError(t, err)
				_, err = ts.models.Transactions.Get(ts.ctx, TransactionFilter{ID: &id})
				assert.ErrorIs(t, err, ErrDocumentNotFound)
			}
		})
	}
}
