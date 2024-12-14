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

const (
	Teacher PatronCategoryType = "teacher"
	Student PatronCategoryType = "student"
)

type PatronCategoryType string

type Patron struct {
	ID        string         `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string         `bson:"name" json:"name"`
	Email     string         `bson:"email" json:"email"`
	CreatedAt time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time      `bson:"updated_at" json:"updated_at"`
	Category  PatronCategory `bson:"category" json:"category"`
	Version   int32          `bson:"version" json:"-"`
}

type PatronFilter struct {
	ID           *string        `json:"id,omitempty"`
	MinCreatedAt *time.Time     `json:"min_created_at,omitempty"`
	MaxCreatedAt *time.Time     `json:"max_created_at,omitempty"`
	MinUpdatedAt *time.Time     `json:"min_updated_at,omitempty"`
	MaxUpdatedAt *time.Time     `json:"max_updated_at,omitempty"`
	Name         *string        `json:"name,omitempty"`
	Email        *string        `json:"email,omitempty"`
	Category     PatronCategory `json:"category,omitempty"`
	Version      *int32         `json:"version,omitempty"`
}

type PatronModel struct {
	Client     *mongo.Client
	Database   string
	Collection string
}

type PatronCategory interface {
	Discount() float64
	Type() PatronCategoryType
}

type TeacherCategory struct {
	CategoryType       PatronCategoryType `bson:"type" json:"type"`
	DiscountPercentage float64            `bson:"discount_percentage" json:"discount_percentage"`
}

type StudentCategory struct {
	CategoryType       PatronCategoryType `bson:"type" json:"type"`
	DiscountPercentage float64            `bson:"discount_percentage" json:"discount_percentage"`
}

func (t TeacherCategory) Discount() float64 {
	return t.DiscountPercentage / 100
}

func (t TeacherCategory) Type() PatronCategoryType {
	return t.CategoryType
}

func (s StudentCategory) Discount() float64 {
	return s.DiscountPercentage / 100
}

func (s StudentCategory) Type() PatronCategoryType {
	return s.CategoryType
}

// NewPatron is a constructor for Patron.
func NewPatron(id string, name, email string, category PatronCategory) *Patron {
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
		query[createdAt] = createdAtRange
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
		categoryFilter := bson.M{}
		switch v := filter.Category.(type) {
		case TeacherCategory:
			categoryFilter[typeTag] = v.Type()
			categoryFilter[discountPercentageTag] = v.DiscountPercentage
		case StudentCategory:
			categoryFilter[typeTag] = v.Type()
			categoryFilter[discountPercentageTag] = v.DiscountPercentage
		}
		query[categoryTag] = categoryFilter
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
	}

	updateFields = append(updateFields, bson.E{Key: updatedAtTag, Value: time.Now()})

	update := bson.D{
		{Key: "$set", Value: updateFields},
		{Key: "$inc", Value: bson.D{{Key: versionTag, Value: 1}}},
	}

	return update
}

// Insert inserts a new Patron into the database.
func (p PatronModel) Insert(ctx context.Context, patron *Patron) (string, error) {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	patron.CreatedAt = time.Now()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := coll.InsertOne(ctx, patron)
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

// Get retrieves a single Patron from the database matching an optional filter.
func (p PatronModel) Get(ctx context.Context, filter PatronFilter) (*Patron, error) {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	filterQuery, err := buildPatronFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	patron := &Patron{}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = coll.FindOne(ctx, filterQuery).Decode(patron)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	return patron, nil
}

// GetAll retrieves all mockBooks from the database matching an optional filter and paginator.
func (p PatronModel) GetAll(ctx context.Context, filter PatronFilter, paginator Paginator) ([]Patron, Metadata, error) {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	patrons := make([]Patron, 0)

	findOpt := options.Find().SetLimit(paginator.limit()).SetSkip(paginator.offset())
	filterQuery, err := buildPatronFilter(filter)
	if err != nil {
		return patrons, Metadata{}, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	totalRecords, err := coll.CountDocuments(ctx, filterQuery)
	if err != nil {
		return patrons, Metadata{}, errCreatingQueryFilter
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

	metadata := calculateMetadata(totalRecords, paginator.Page, paginator.PageSize)

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

// Delete deletes a Patron from the database by ID.
func (p PatronModel) Delete(ctx context.Context, filter PatronFilter) error {
	coll := p.Client.Database(p.Database).Collection(p.Collection)

	filterQuery, err := buildPatronFilter(filter)
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
