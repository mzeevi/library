package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/mzeevi/library/internal/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
)

var (
	AnonymousAdmin = &Admin{}
)

type Admin struct {
	ID          string        `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string        `bson:"name" json:"name"`
	Activated   bool          `bson:"activated" json:"activated"`
	Password    auth.Password `bson:"password" json:"-"`
	Permissions []string      `bson:"permissions" json:"-"`
}

type AdminModel struct {
	Client     *mongo.Client
	Database   string
	Collection string
}

type AdminFilter struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// IsAnonymous checks if a Patron instance is anonymous.
func (a *Admin) IsAnonymous() bool {
	return a == AnonymousAdmin
}

// buildAdminFilter constructs a filter query for filtering admins.
func buildAdminFilter(filter AdminFilter) (bson.M, error) {
	query := bson.M{}

	if filter.ID != nil {
		id, err := primitive.ObjectIDFromHex(*filter.ID)
		if err != nil {
			return query, err
		}
		query[idTag] = id
	}

	if filter.Name != nil {
		query[nameTag] = *filter.Name
	}

	return query, nil
}

// generateAdmin constructs a new Admin.
func generateAdmin(username, password string) (*Admin, error) {
	admin := &Admin{}

	admin.Name = username
	admin.Activated = true
	admin.Permissions = auth.AdminPermissions

	err := admin.Password.Set(password)
	if err != nil {
		return admin, err
	}

	return admin, nil
}

func (a AdminModel) New(ctx context.Context, username, password string) error {
	admin, err := generateAdmin(username, password)
	if err != nil {
		return err
	}

	return a.Insert(ctx, admin)
}

// Insert inserts a new Patron into the database.
func (a AdminModel) Insert(ctx context.Context, admin *Admin) error {
	coll := a.Client.Database(a.Database).Collection(a.Collection)

	_, err := coll.InsertOne(ctx, admin)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "_id_ dup key:"):
			return ErrDuplicateID
		default:
			return err
		}
	}

	return nil
}

// Get retrieves a single Patron from the database matching an optional filter.
func (a AdminModel) Get(ctx context.Context, filter AdminFilter) (*Admin, error) {
	coll := a.Client.Database(a.Database).Collection(a.Collection)

	filterQuery, err := buildAdminFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	admin := &Admin{}

	err = coll.FindOne(ctx, filterQuery).Decode(admin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	return admin, nil
}
