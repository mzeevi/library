package data

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrDuplicateID      = errors.New("duplicate id")
	ErrEditConflict     = errors.New("edit conflict")
)

const (
	BooksCollectionKey        = "books"
	PatronsCollectionKey      = "patrons"
	TransactionsCollectionKey = "transactions"
	TokensCollectionKey       = "tokens"
	AdminsCollectionKey       = "admins"
)

type Models struct {
	Books        BookModel
	Patrons      PatronModel
	Transactions TransactionModel
	Tokens       TokenModel
	Admins       AdminModel
}

func NewModels(client *mongo.Client, database string, collections map[string]string) Models {
	return Models{
		Books:        BookModel{Client: client, Database: database, Collection: collections[BooksCollectionKey]},
		Patrons:      PatronModel{Client: client, Database: database, Collection: collections[PatronsCollectionKey]},
		Transactions: TransactionModel{Client: client, Database: database, Collection: collections[TransactionsCollectionKey]},
		Tokens:       TokenModel{Client: client, Database: database, Collection: collections[TokensCollectionKey]},
		Admins:       AdminModel{Client: client, Database: database, Collection: collections[AdminsCollectionKey]},
	}
}
