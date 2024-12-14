package data

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrDuplicateID      = errors.New("duplicate id")
	ErrEditConflict     = errors.New("edit conflict")
)

var (
	timeout = 10 * time.Second
)

const (
	idTag        = "_id"
	createdAt    = "created_at"
	updatedAtTag = "updated_at"
)

const (
	patronIDTag           = "patron_id"
	bookIDTag             = "book_id"
	borrowedAtTag         = "borrowed_at"
	dueDateTag            = "due_date"
	returnedAtTag         = "returned_at"
	statusTag             = "status"
	typeTag               = "type"
	discountPercentageTag = "discount_percentage"
)

const (
	pagesTag          = "pages"
	editionTag        = "edition"
	publishedAtTag    = "published_at"
	createdAtTag      = "created_at"
	titleTag          = "title"
	isbnTag           = "isbn"
	authorsTag        = "authors"
	publishersTag     = "publishers"
	genresTag         = "genres"
	versionTag        = "version"
	copiesTag         = "copies"
	borrowedCopiesTag = "borrowed_copies"
)

const (
	nameTag     = "name"
	emailTag    = "email"
	categoryTag = "category"
)

const (
	BooksCollectionKey        = "books"
	PatronsCollectionKey      = "patrons"
	TransactionsCollectionKey = "transactions"
)

type Models struct {
	Books        BookModel
	Patrons      PatronModel
	Transactions TransactionModel
}

func NewModels(client *mongo.Client, database string, collections map[string]string) Models {
	return Models{
		Books:        BookModel{Client: client, Database: database, Collection: collections[BooksCollectionKey]},
		Patrons:      PatronModel{Client: client, Database: database, Collection: collections[PatronsCollectionKey]},
		Transactions: TransactionModel{Client: client, Database: database, Collection: collections[TransactionsCollectionKey]},
	}
}
