package data

import (
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

const (
	errUnknownCategory = "unknown patron category"
)

// MarshalBSON customizes the BSON encoding for the Patron struct.
// It serializes the Patron struct into BSON, handling the Category field properly,
// converting it to the appropriate type string ("teacher" or "student") and attaching
// the discount percentage.
func (p *Patron) MarshalBSON() ([]byte, error) {
	var categoryType string
	var discountPercentage float64

	switch c := p.Category.(type) {
	case TeacherCategory:
		categoryType = string(Teacher)
		discountPercentage = c.DiscountPercentage
	case StudentCategory:
		categoryType = string(Student)
		discountPercentage = c.DiscountPercentage
	default:
		return []byte{}, errors.New(errUnknownCategory)
	}

	aux := struct {
		ID        string    `bson:"_id,omitempty"`
		Name      string    `bson:"name"`
		Email     string    `bson:"email"`
		CreatedAt time.Time `bson:"created_at"`
		Category  struct {
			Type               string  `bson:"type"`
			DiscountPercentage float64 `bson:"discount_percentage"`
		} `bson:"category"`
		Version int32 `bson:"version"`
	}{
		ID:        p.ID,
		Name:      p.Name,
		Email:     p.Email,
		CreatedAt: p.CreatedAt,
		Category: struct {
			Type               string  `bson:"type"`
			DiscountPercentage float64 `bson:"discount_percentage"`
		}{
			Type:               categoryType,
			DiscountPercentage: discountPercentage,
		},
		Version: p.Version,
	}

	return bson.Marshal(aux)
}

// UnmarshalBSON customizes the BSON decoding for the Patron struct.
// It deserializes the BSON data into the Patron struct and handles the Category field,
// which can be either a TeacherCategory or a StudentCategory, depending on the type in BSON.
func (p *Patron) UnmarshalBSON(data []byte) error {
	aux := struct {
		ID        string    `bson:"_id,omitempty"`
		Name      string    `bson:"name"`
		Email     string    `bson:"email"`
		CreatedAt time.Time `bson:"created_at"`
		Category  struct {
			Type               string  `bson:"type"`
			DiscountPercentage float64 `bson:"discount_percentage"`
		} `bson:"category"`
		Version int32 `bson:"version"`
	}{}

	if err := bson.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.ID = aux.ID
	p.Name = aux.Name
	p.Email = aux.Email
	p.CreatedAt = aux.CreatedAt
	p.Version = aux.Version

	switch aux.Category.Type {
	case string(Teacher):
		p.Category = TeacherCategory{
			CategoryType:       Teacher,
			DiscountPercentage: aux.Category.DiscountPercentage,
		}
	case string(Student):
		p.Category = StudentCategory{
			CategoryType:       Student,
			DiscountPercentage: aux.Category.DiscountPercentage,
		}
	default:
		return errors.New(errUnknownCategory)
	}

	return nil
}

// MarshalJSON customizes the JSON encoding for the Patron struct.
// It serializes the Patron struct into JSON, ensuring the Category field is handled properly,
// using the correct string representation ("teacher" or "student") and including the discount percentage.
func (p *Patron) MarshalJSON() ([]byte, error) {
	var categoryType string
	var discountPercentage float64

	switch c := p.Category.(type) {
	case TeacherCategory:
		categoryType = string(Teacher)
		discountPercentage = c.DiscountPercentage
	case StudentCategory:
		categoryType = string(Student)
		discountPercentage = c.DiscountPercentage
	default:
		return []byte{}, errors.New(errUnknownCategory)
	}

	aux := struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		Category  struct {
			Type               string  `json:"type"`
			DiscountPercentage float64 `json:"discount_percentage"`
		} `json:"category"`
		Version int32 `json:"version"`
	}{
		ID:        p.ID,
		Name:      p.Name,
		Email:     p.Email,
		CreatedAt: p.CreatedAt,
		Category: struct {
			Type               string  `json:"type"`
			DiscountPercentage float64 `json:"discount_percentage"`
		}{
			Type:               categoryType,
			DiscountPercentage: discountPercentage,
		},
		Version: p.Version,
	}

	return json.Marshal(aux)
}

// UnmarshalJSON customizes the JSON decoding for the Patron struct.
// It deserializes the JSON data into the Patron struct, handling the Category field,
// and ensuring the Category is either a TeacherCategory or a StudentCategory.
func (p *Patron) UnmarshalJSON(data []byte) error {
	aux := struct {
		ID        string    `json:"id,omitempty"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		Category  struct {
			Type               string  `json:"type"`
			DiscountPercentage float64 `json:"discount_percentage"`
		} `json:"category"`
		Version int32 `json:"version"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.ID = aux.ID
	p.Name = aux.Name
	p.Email = aux.Email
	p.CreatedAt = aux.CreatedAt
	p.Version = aux.Version

	switch aux.Category.Type {
	case string(Teacher):
		p.Category = TeacherCategory{
			CategoryType:       Teacher,
			DiscountPercentage: aux.Category.DiscountPercentage,
		}
	case string(Student):
		p.Category = StudentCategory{
			CategoryType:       Student,
			DiscountPercentage: aux.Category.DiscountPercentage,
		}
	default:
		return errors.New(errUnknownCategory)
	}

	return nil
}
