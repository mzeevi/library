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

type Book struct {
	ID             string    `bson:"_id,omitempty" json:"id,omitempty"`
	Pages          int       `bson:"pages" json:"pages"`
	Edition        int       `bson:"edition" json:"edition"`
	Copies         int       `bson:"copies" json:"copies"`
	BorrowedCopies int       `bson:"borrowed_copies" json:"borrowed_copies"`
	PublishedAt    time.Time `bson:"published_at" json:"published_at"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
	Title          string    `bson:"title" json:"title"`
	ISBN           string    `bson:"isbn" json:"isbn"`
	Authors        []string  `bson:"authors" json:"authors"`
	Publishers     []string  `bson:"publishers" json:"publishers"`
	Genres         []string  `bson:"genres" json:"genres"`
	Version        int32     `bson:"version" json:"-"`
}

type BookFilter struct {
	ID                *string    `json:"id,omitempty"`
	MinPages          *int       `json:"min_pages,omitempty"`
	MaxPages          *int       `json:"max_pages,omitempty"`
	MinEdition        *int       `json:"min_edition,omitempty"`
	MaxEdition        *int       `json:"max_edition,omitempty"`
	MinPublishedAt    *time.Time `json:"min_published_at,omitempty"`
	MaxPublishedAt    *time.Time `json:"max_published_at,omitempty"`
	MinCreatedAt      *time.Time `json:"min_created_at,omitempty"`
	MaxCreatedAt      *time.Time `json:"max_created_at,omitempty"`
	MinUpdatedAt      *time.Time `json:"min_updated_at,omitempty"`
	MaxUpdatedAt      *time.Time `json:"max_updated_at,omitempty"`
	Title             *string    `json:"title,omitempty"`
	ISBN              *string    `json:"isbn,omitempty"`
	Authors           []string   `json:"authors,omitempty"`
	Publishers        []string   `json:"publishers,omitempty"`
	Genres            []string   `json:"genres,omitempty"`
	Version           *int32     `json:"version,omitempty"`
	MinCopies         *int       `json:"min_copies,omitempty"`
	MaxCopies         *int       `json:"max_copies,omitempty"`
	MinBorrowedCopies *int       `json:"min_borrowed_copies,omitempty"`
	MaxBorrowedCopies *int       `json:"max_borrowed_copies,omitempty"`
}

type BookModel struct {
	Client     *mongo.Client
	Database   string
	Collection string
}

// NewBook creates a new book with the provided details.
func NewBook(id string, title, isbn string, pages, edition, copies int, authors, publishers, genres []string, publishedAt time.Time) *Book {
	now := time.Now()
	return &Book{
		ID:          id,
		Title:       title,
		ISBN:        isbn,
		Copies:      copies,
		Pages:       pages,
		Edition:     edition,
		Authors:     authors,
		Publishers:  publishers,
		Genres:      genres,
		PublishedAt: publishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// buildBookFilter constructs a filter query for filtering books.
func buildBookFilter(filter BookFilter) (bson.M, error) {
	query := bson.M{}

	if filter.ID != nil {
		id, err := primitive.ObjectIDFromHex(*filter.ID)
		if err != nil {
			return query, err
		}
		query[idTag] = id
	}
	if filter.MinPages != nil || filter.MaxPages != nil {
		pagesRange := bson.M{}
		if filter.MinPages != nil {
			pagesRange["$gte"] = *filter.MinPages
		}
		if filter.MaxPages != nil {
			pagesRange["$lte"] = *filter.MaxPages
		}
		query[pagesTag] = pagesRange
	}
	if filter.MinEdition != nil || filter.MaxEdition != nil {
		editionRange := bson.M{}
		if filter.MinEdition != nil {
			editionRange["$gte"] = *filter.MinEdition
		}
		if filter.MaxEdition != nil {
			editionRange["$lte"] = *filter.MaxEdition
		}
		query[editionTag] = editionRange
	}
	if filter.MinPublishedAt != nil || filter.MaxPublishedAt != nil {
		publishedAtRange := bson.M{}
		if filter.MinPublishedAt != nil {
			publishedAtRange["$gte"] = *filter.MinPublishedAt
		}
		if filter.MaxPublishedAt != nil {
			publishedAtRange["$lte"] = *filter.MaxPublishedAt
		}
		query[publishedAtTag] = publishedAtRange
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
	if filter.Title != nil {
		query[titleTag] = bson.M{"$regex": *filter.Title, "$options": "i"}
	}
	if filter.ISBN != nil {
		query[isbnTag] = *filter.ISBN
	}
	if len(filter.Authors) > 0 {
		query[authorsTag] = bson.M{"$in": filter.Authors}
	}
	if len(filter.Publishers) > 0 {
		query[publishersTag] = bson.M{"$in": filter.Publishers}
	}
	if len(filter.Genres) > 0 {
		query[genresTag] = bson.M{"$in": filter.Genres}
	}
	if filter.Version != nil {
		query[versionTag] = *filter.Version
	}
	if filter.MinCopies != nil || filter.MaxCopies != nil {
		copiesRange := bson.M{}
		if filter.MinCopies != nil {
			copiesRange["$gte"] = *filter.MinCopies
		}
		if filter.MaxCopies != nil {
			copiesRange["$lte"] = *filter.MaxCopies
		}
		query[copiesTag] = copiesRange
	}
	if filter.MinBorrowedCopies != nil || filter.MaxBorrowedCopies != nil {
		borrowedCopiesRange := bson.M{}
		if filter.MinBorrowedCopies != nil {
			borrowedCopiesRange["$gte"] = *filter.MinBorrowedCopies
		}
		if filter.MaxBorrowedCopies != nil {
			borrowedCopiesRange["$lte"] = *filter.MaxBorrowedCopies
		}
		query[borrowedCopiesTag] = borrowedCopiesRange
	}

	return query, nil
}

// buildBookUpdater constructs an update document for updating a Book.
func buildBookUpdater(book *Book) bson.D {
	updateFields := bson.D{
		{Key: titleTag, Value: book.Title},
		{Key: isbnTag, Value: book.ISBN},
		{Key: pagesTag, Value: book.Pages},
		{Key: editionTag, Value: book.Edition},
		{Key: publishedAtTag, Value: book.PublishedAt},
		{Key: authorsTag, Value: book.Authors},
		{Key: publishersTag, Value: book.Publishers},
		{Key: genresTag, Value: book.Genres},
		{Key: copiesTag, Value: book.Copies},
		{Key: borrowedCopiesTag, Value: book.BorrowedCopies},
	}

	updateFields = append(updateFields, bson.E{Key: updatedAtTag, Value: time.Now()})

	update := bson.D{
		{Key: "$set", Value: updateFields},
		{Key: "$inc", Value: bson.D{{Key: versionTag, Value: 1}}},
	}

	return update
}

// CreateUniqueISBNIndex creates a unique index using the ISBN field.
func (b BookModel) CreateUniqueISBNIndex() error {
	coll := b.Client.Database(b.Database).Collection(b.Collection)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: isbnTag, Value: -1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new Book into the database.
func (b BookModel) Insert(ctx context.Context, book *Book) (string, error) {
	coll := b.Client.Database(b.Database).Collection(b.Collection)

	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := coll.InsertOne(ctx, book)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "E11000 duplicate key error collection"):
			return "", ErrDuplicateID
		default:
			return "", err
		}
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return res.InsertedID.(string), nil
}

// Get retrieves a single Book from the database matching an optional filter.
func (b BookModel) Get(ctx context.Context, filter BookFilter) (*Book, error) {
	coll := b.Client.Database(b.Database).Collection(b.Collection)

	filterQuery, err := buildBookFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	book := &Book{}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = coll.FindOne(ctx, filterQuery).Decode(book)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	return book, nil
}

// GetAll retrieves all mockBooks from the database matching an optional filter and paginator.
func (b BookModel) GetAll(ctx context.Context, filter BookFilter, paginator Paginator) ([]Book, Metadata, error) {
	coll := b.Client.Database(b.Database).Collection(b.Collection)

	books := make([]Book, 0)

	findOpt := options.Find().SetLimit(paginator.limit()).SetSkip(paginator.offset())
	filterQuery, err := buildBookFilter(filter)
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	totalRecords, err := coll.CountDocuments(ctx, filterQuery)
	if err != nil {
		return books, Metadata{}, errCreatingQueryFilter
	}

	cursor, err := coll.Find(ctx, filterQuery, findOpt)
	if err != nil {
		return books, Metadata{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var book Book
		if err = cursor.Decode(&book); err != nil {
			return books, Metadata{}, err
		}

		books = append(books, book)
	}

	metadata := calculateMetadata(totalRecords, paginator.Page, paginator.PageSize)

	return books, metadata, nil
}

// Update updates a Book's details in the database.
func (b BookModel) Update(ctx context.Context, filter BookFilter, book *Book) error {
	coll := b.Client.Database(b.Database).Collection(b.Collection)

	update := buildBookUpdater(book)

	filter.Version = &book.Version
	filterQuery, err := buildBookFilter(filter)
	if err != nil {
		return fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := coll.UpdateOne(ctx, filterQuery, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrEditConflict
	}

	return nil
}

// Delete deletes a Book from the database by ID.
func (b BookModel) Delete(ctx context.Context, filter BookFilter) error {
	coll := b.Client.Database(b.Database).Collection(b.Collection)

	filterQuery, err := buildBookFilter(filter)
	if err != nil {
		return fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := coll.DeleteOne(ctx, filterQuery)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}
