package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `bson:"plaintext" json:"token,omitempty"`
	Hash      []byte    `bson:"hash" json:"-"`
	PatronID  string    `bson:"patron_id" json:"-"`
	Expiry    time.Time `bson:"expiry" json:"expiry,omitempty"`
	Scope     string    `bson:"scope" json:"-,omitempty"`
}

type TokenModel struct {
	Client     *mongo.Client
	Database   string
	Collection string
}

type TokenFilter struct {
	Plaintext *string
	PatronID  *string
	Hash      []byte
	MinExpiry *time.Time
	MaxExpiry *time.Time
	Scope     *string
}

// buildTokenFilter constructs a filter query for filtering tokens.
func buildTokenFilter(filter TokenFilter) (bson.M, error) {
	query := bson.M{}

	if filter.PatronID != nil {
		query[patronIDTag] = *filter.PatronID
	}

	if filter.Plaintext != nil {
		query[plaintextTag] = *filter.Plaintext
	}

	if len(filter.Hash) > 0 {
		query[hashTag] = filter.Hash
	}

	if filter.MinExpiry != nil || filter.MaxExpiry != nil {
		expiryRange := bson.M{}
		if filter.MinExpiry != nil {
			expiryRange["$gte"] = *filter.MinExpiry
		}
		if filter.MaxExpiry != nil {
			expiryRange["$lte"] = *filter.MaxExpiry
		}
		query[expiryTag] = expiryRange
	}

	if filter.Scope != nil {
		query[scopeTag] = *filter.Scope
	}

	return query, nil
}

// generateToken creates a new Token for the specified patron, with a given time-to-live (TTL)
// and scope. The Token includes a unique plaintext identifier and an SHA-256 hash of that plaintext.
func generateToken(patronID string, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		PatronID: patronID,
		Expiry:   time.Now().Add(ttl),
		Scope:    scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func (t TokenModel) New(ctx context.Context, patronID string, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(patronID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = t.Insert(ctx, token)
	return token, err
}

// Insert inserts a new Token into the database.
func (t TokenModel) Insert(ctx context.Context, token *Token) error {
	coll := t.Client.Database(t.Database).Collection(t.Collection)

	_, err := coll.InsertOne(ctx, token)
	return err
}

// GetPatronID returns the patronID which matches the filter.
func (t TokenModel) GetPatronID(ctx context.Context, filter TokenFilter) (string, error) {
	coll := t.Client.Database(t.Database).Collection(t.Collection)

	filterQuery, err := buildTokenFilter(filter)
	if err != nil {
		return "", fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	token := &Token{}

	err = coll.FindOne(ctx, filterQuery).Decode(&token)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrDocumentNotFound
		}
		return "", err
	}

	return token.PatronID, nil
}

// DeleteAllForPatron deletes multiple Tokens from the database by filter.
func (t TokenModel) DeleteAllForPatron(ctx context.Context, filter TokenFilter) error {
	coll := t.Client.Database(t.Database).Collection(t.Collection)

	filterQuery, err := buildTokenFilter(filter)
	if err != nil {
		return fmt.Errorf("%v: %v", errCreatingQueryFilter, err)
	}

	result, err := coll.DeleteMany(ctx, filterQuery)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}
