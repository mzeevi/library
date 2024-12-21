package data

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

var (
	errCreatingQueryFilter = errors.New("filter query builder failed")
	errCreatingQuerySort   = errors.New("sort query builder failed")
)

const (
	TransactionStatusBorrowed = "borrowed"
	TransactionStatusReturned = "returned"
)

type Transaction struct {
	ID         string    `bson:"_id,omitempty" json:"id,omitempty"`
	PatronID   string    `bson:"patron_id" json:"patron_id"`
	BookID     string    `bson:"book_id" json:"book_id"`
	Status     string    `bson:"status" json:"status"`
	BorrowedAt time.Time `bson:"borrowed_at" json:"borrowed_at"`
	DueDate    time.Time `bson:"due_date" json:"due_date"`
	ReturnedAt time.Time `bson:"returned_at,omitempty" json:"returned_at,omitempty"`
	CreatedAt  time.Time `bson:"created_at" json:"-"`
	UpdatedAt  time.Time `bson:"updated_at" json:"-"`
	Version    int32     `bson:"version" json:"-"`
}

type TransactionFilter struct {
	ID            *string    `json:"id,omitempty"`
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
	MinUpdatedAt  *time.Time `json:"min_updated_at,omitempty"`
	MaxUpdatedAt  *time.Time `json:"max_updated_at,omitempty"`
	Version       *int32     `json:"-,omitempty"`
}

type TransactionModel struct {
	Client     *mongo.Client
	Database   string
	Collection string
}

// NewTransaction is a constructor for Transaction.
func NewTransaction(id, patronID, bookID, status string, borrowedAt, dueDate time.Time) *Transaction {
	now := time.Now()
	return &Transaction{
		ID:         id,
		PatronID:   patronID,
		BookID:     bookID,
		BorrowedAt: borrowedAt,
		DueDate:    dueDate,
		Status:     status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// buildTransactionFilter constructs a filter query for filtering transactions.
func buildTransactionFilter(filter TransactionFilter) (bson.M, error) {
	query := bson.M{}

	if filter.ID != nil {
		id, err := primitive.ObjectIDFromHex(*filter.ID)
		if err != nil {
			return query, err
		}
		query[idTag] = id
	}
	if filter.PatronID != nil {
		query[patronIDTag] = *filter.PatronID
	}
	if filter.BookID != nil {
		query[bookIDTag] = *filter.BookID
	}
	if filter.Status != nil {
		query[statusTag] = *filter.Status
	}

	if filter.MinReturnedAt != nil || filter.MaxReturnedAt != nil {
		returnedAtRange := bson.M{}
		if filter.MinReturnedAt != nil {
			returnedAtRange["$gte"] = *filter.MinReturnedAt
		}
		if filter.MaxReturnedAt != nil {
			returnedAtRange["$lte"] = *filter.MaxReturnedAt
		}
		query[borrowedAtTag] = returnedAtRange
	}

	if filter.MinBorrowedAt != nil || filter.MaxBorrowedAt != nil {
		borrowedAtRange := bson.M{}
		if filter.MinBorrowedAt != nil {
			borrowedAtRange["$gte"] = *filter.MinBorrowedAt
		}
		if filter.MaxBorrowedAt != nil {
			borrowedAtRange["$lte"] = *filter.MaxBorrowedAt
		}
		query[borrowedAtTag] = borrowedAtRange
	}

	if filter.MinDueDate != nil || filter.MaxDueDate != nil {
		dueDateRange := bson.M{}
		if filter.MinDueDate != nil {
			dueDateRange["$gte"] = *filter.MinDueDate
		}
		if filter.MaxDueDate != nil {
			dueDateRange["$lte"] = *filter.MaxDueDate
		}
		query[dueDateTag] = dueDateRange
	}

	if filter.MinCreatedAt != nil || filter.MaxCreatedAt != nil {
		createdAtRange := bson.M{}
		if filter.MinCreatedAt != nil {
			createdAtRange["$gte"] = *filter.MinCreatedAt
		}
		if filter.MaxCreatedAt != nil {
			createdAtRange["$lte"] = *filter.MaxCreatedAt
		}
		query[createdAtTag] = createdAtRange
	}

	if filter.MinUpdatedAt != nil || filter.MaxUpdatedAt != nil {
		updatedAtRange := bson.M{}
		if filter.MinUpdatedAt != nil {
			updatedAtRange["$gte"] = *filter.MinUpdatedAt
		}
		if filter.MaxUpdatedAt != nil {
			updatedAtRange["$lte"] = *filter.MaxUpdatedAt
		}
		query[updatedAtTag] = updatedAtRange
	}

	if filter.Version != nil {
		query[versionTag] = *filter.Version
	}

	return query, nil
}

// buildTransactionUpdater constructs an update document for updating a Transaction.
func buildTransactionUpdater(transaction *Transaction) bson.D {
	updateFields := bson.D{
		{Key: dueDateTag, Value: transaction.DueDate},
		{Key: returnedAtTag, Value: transaction.ReturnedAt},
		{Key: statusTag, Value: transaction.Status},
	}

	updateFields = append(updateFields, bson.E{Key: updatedAtTag, Value: time.Now()})

	update := bson.D{
		{Key: "$set", Value: updateFields},
		{Key: "$inc", Value: bson.D{{Key: versionTag, Value: 1}}},
	}

	return update
}

// Insert inserts a new Transaction into the database.
func (t TransactionModel) Insert(ctx context.Context, transaction *Transaction) (string, error) {
	coll := t.Client.Database(t.Database).Collection(t.Collection)

	now := time.Now()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	res, err := coll.InsertOne(ctx, transaction)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "E11000 duplicate key error collection"):
			return "", ErrDuplicateID
		default:
			return "", err
		}
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Get retrieves a single Transaction from the database matching an optional filter.
func (t TransactionModel) Get(ctx context.Context, filter TransactionFilter) (*Transaction, error) {
	coll := t.Client.Database(t.Database).Collection(t.Collection)

	filterQuery, err := buildTransactionFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	transaction := &Transaction{}

	if err = coll.FindOne(ctx, filterQuery).Decode(transaction); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	return transaction, nil
}

// GetAll retrieves all Transactions from the database matching an optional filter and paginator.
func (t TransactionModel) GetAll(ctx context.Context, filter TransactionFilter, paginator Paginator, sorter Sorter) ([]Transaction, Metadata, error) {
	coll := t.Client.Database(t.Database).Collection(t.Collection)

	transactions := make([]Transaction, 0)
	metadata := Metadata{}

	filterQuery, err := buildTransactionFilter(filter)
	if err != nil {
		return transactions, Metadata{}, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	sortQuery, err := buildSorter(sorter)
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("%v: %v", errCreatingQuerySort, err)
	}

	findOpt := options.Find().SetSort(sortQuery)

	if paginator.valid() {
		var totalRecords int64

		findOpt = findOpt.SetLimit(paginator.limit()).SetSkip(paginator.offset())
		totalRecords, err = coll.CountDocuments(ctx, filterQuery)
		if err != nil {
			return transactions, Metadata{}, errCreatingQueryFilter
		}

		metadata = calculateMetadata(totalRecords, paginator.Page, paginator.PageSize)
	}

	cursor, err := coll.Find(ctx, filterQuery, findOpt)
	if err != nil {
		return transactions, Metadata{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var transaction Transaction
		if err = cursor.Decode(&transaction); err != nil {
			return transactions, Metadata{}, err
		}

		transactions = append(transactions, transaction)
	}

	if err = cursor.Err(); err != nil {
		return transactions, Metadata{}, err
	}

	return transactions, metadata, nil
}

// Update updates a Transaction's details in the database.
func (t TransactionModel) Update(ctx context.Context, filter TransactionFilter, transaction *Transaction) error {
	coll := t.Client.Database(t.Database).Collection(t.Collection)

	update := buildTransactionUpdater(transaction)

	filter.Version = &transaction.Version
	filterQuery, err := buildTransactionFilter(filter)
	if err != nil {
		return fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	result, err := coll.UpdateOne(ctx, filterQuery, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrEditConflict
	}

	return nil
}

// Delete deletes a Transaction from the database by ID.
func (t TransactionModel) Delete(ctx context.Context, filter TransactionFilter) error {
	coll := t.Client.Database(t.Database).Collection(t.Collection)

	filterQuery, err := buildTransactionFilter(filter)
	if err != nil {
		return fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	result, err := coll.DeleteOne(ctx, filterQuery)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}
