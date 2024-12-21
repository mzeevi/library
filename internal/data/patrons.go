package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/mzeevi/library/internal/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var (
	AnonymousPatron = &Patron{}
)

type Patron struct {
	ID          string        `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string        `bson:"name" json:"name"`
	Email       string        `bson:"email" json:"email"`
	Category    string        `bson:"category" json:"category"`
	Password    auth.Password `bson:"password" json:"-"`
	Activated   bool          `bson:"activated" json:"activated"`
	Permissions []string      `bson:"permissions" json:"-"`
	Version     int32         `bson:"version" json:"-"`
	CreatedAt   time.Time     `bson:"created_at" json:"-"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"-"`
}

type PatronFilter struct {
	ID           *string    `json:"id,omitempty"`
	Name         *string    `json:"name,omitempty"`
	Email        *string    `json:"email,omitempty"`
	Category     *string    `json:"category,omitempty"`
	Version      *int32     `json:"version,omitempty"`
	MinCreatedAt *time.Time `json:"min_created_at,omitempty"`
	MaxCreatedAt *time.Time `json:"max_created_at,omitempty"`
	MinUpdatedAt *time.Time `json:"min_updated_at,omitempty"`
	MaxUpdatedAt *time.Time `json:"max_updated_at,omitempty"`
}

type PatronModel struct {
	Client     *mongo.Client
	Database   string
	Collection string
}

// NewPatron is a constructor for Patron.
func NewPatron(id string, name, email, category string) *Patron {
	now := time.Now()

	return &Patron{
		ID:        id,
		Name:      name,
		Email:     email,
		Category:  category,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsAnonymous checks if a Patron instance is anonymous.
func (p *Patron) IsAnonymous() bool {
	return p == AnonymousPatron
}

// buildPatronFilter constructs a filter query for filtering patrons.
func buildPatronFilter(filter PatronFilter) (bson.M, error) {
	query := bson.M{}

	if filter.ID != nil {
		id, err := primitive.ObjectIDFromHex(*filter.ID)
		if err != nil {
			return query, err
		}
		query[idTag] = id
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
	if filter.Name != nil {
		query[nameTag] = bson.M{"$regex": *filter.Name, "$options": "i"}
	}
	if filter.Email != nil {
		query[emailTag] = *filter.Email
	}
	if filter.Category != nil {
		query[categoryTag] = *filter.Category
	}
	if filter.Version != nil {
		query[versionTag] = *filter.Version
	}

	return query, nil
}

// buildPatronUpdater constructs an update document for updating a Patron.
func buildPatronUpdater(patron *Patron) bson.D {
	updateFields := bson.D{
		{Key: nameTag, Value: patron.Name},
		{Key: emailTag, Value: patron.Email},
		{Key: categoryTag, Value: patron.Category},
		{Key: passwordTag, Value: patron.Password},
		{Key: activatedTag, Value: patron.Activated},
		{Key: permissionsTag, Value: patron.Permissions},
	}

	updateFields = append(updateFields, bson.E{Key: updatedAtTag, Value: time.Now()})

	update := bson.D{
		{Key: "$set", Value: updateFields},
		{Key: "$inc", Value: bson.D{{Key: versionTag, Value: 1}}},
	}

	return update
}

// CreateUniqueIndex creates a unique index using a field.
func (p PatronModel) CreateUniqueIndex() error {
	coll := p.Client.Database(p.Database).Collection(p.Collection)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: emailTag, Value: -1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new Patron into the database.
func (p PatronModel) Insert(ctx context.Context, patron *Patron) (string, error) {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	patron.CreatedAt = time.Now()

	res, err := coll.InsertOne(ctx, patron)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "_id_ dup key:"):
			return "", ErrDuplicateID
		case strings.Contains(err.Error(), "email_-1 dup key"):
			return "", ErrDuplicateEmail
		default:
			return "", err
		}
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return res.InsertedID.(string), nil
}

// Get retrieves a single Patron from the database matching an optional filter.
func (p PatronModel) Get(ctx context.Context, filter PatronFilter) (*Patron, error) {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	filterQuery, err := buildPatronFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	patron := &Patron{}

	err = coll.FindOne(ctx, filterQuery).Decode(patron)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	return patron, nil
}

// GetAll retrieves all patrons from the database matching an optional filter and paginator.
func (p PatronModel) GetAll(ctx context.Context, filter PatronFilter, paginator Paginator, sorter Sorter) ([]Patron, Metadata, error) {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	patrons := make([]Patron, 0)
	metadata := Metadata{}

	filterQuery, err := buildPatronFilter(filter)
	if err != nil {
		return patrons, Metadata{}, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	sortQuery, err := buildSorter(sorter)
	if err != nil {
		return patrons, Metadata{}, fmt.Errorf("%v: %v", errCreatingQuerySort, err)
	}

	findOpt := options.Find().SetSort(sortQuery)

	if paginator.valid() {
		var totalRecords int64

		findOpt = findOpt.SetLimit(paginator.limit()).SetSkip(paginator.offset())
		totalRecords, err = coll.CountDocuments(ctx, filterQuery)
		if err != nil {
			return patrons, Metadata{}, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
		}

		metadata = calculateMetadata(totalRecords, paginator.Page, paginator.PageSize)
	}

	cursor, err := coll.Find(ctx, filterQuery, findOpt)
	if err != nil {
		return patrons, Metadata{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var patron Patron
		if err = cursor.Decode(&patron); err != nil {
			return patrons, Metadata{}, err
		}

		patrons = append(patrons, patron)
	}

	if err = cursor.Err(); err != nil {
		return patrons, Metadata{}, err
	}

	return patrons, metadata, nil
}

// Update updates a Patron's details in the database.
func (p PatronModel) Update(ctx context.Context, filter PatronFilter, patron *Patron) error {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	update := buildPatronUpdater(patron)

	filter.Version = &patron.Version
	filterQuery, err := buildPatronFilter(filter)
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

// Delete deletes a Patron from the database by filter.
func (p PatronModel) Delete(ctx context.Context, filter PatronFilter) error {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	filterQuery, err := buildPatronFilter(filter)
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
