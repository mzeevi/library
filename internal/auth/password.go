package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Plaintext *string `bson:"plaintext" json:"plaintext"`
	Hash      []byte  `bson:"hash" json:"-"`
}

// Set calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *Password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.Plaintext = &plaintextPassword
	p.Hash = hash

	return nil
}

// Matches checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false otherwise.
func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
