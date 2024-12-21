package data

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

type Paginator struct {
	Page     int64
	PageSize int64
}

type Sorter struct {
	Field        string
	SortSafelist []string
}

type Metadata struct {
	CurrentPage  int64 `json:"current_page,omitempty"`
	PageSize     int64 `json:"page_size,omitempty"`
	FirstPage    int64 `json:"first_page,omitempty"`
	LastPage     int64 `json:"last_page,omitempty"`
	TotalRecords int64 `json:"total_records,omitempty"`
}

func (p Paginator) valid() bool {
	return p.Page > 0 && p.PageSize > 0
}

func (p Paginator) limit() int64 {
	return p.PageSize
}

func (p Paginator) offset() int64 {
	return (p.Page - 1) * p.PageSize
}

func (s Sorter) field() (string, error) {
	if s.Field == "" {
		return "", nil
	}

	for _, safeValue := range s.SortSafelist {
		if s.Field == safeValue {
			return strings.TrimPrefix(s.Field, "-"), nil
		}
	}

	return "", fmt.Errorf("unsupported sort field")
}

func (s Sorter) sortDirection() int {
	if strings.HasPrefix(s.Field, "-") {
		return -1
	}

	return 1
}

// calculateMetadata returns metadata regarding pagination.
func calculateMetadata(totalRecords, page, pageSize int64) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}

// buildSorter constructs a sort query for sorting.
func buildSorter(sorter Sorter) (bson.D, error) {
	query := bson.D{}

	field, err := sorter.field()
	if err != nil {
		return query, err
	} else if field == "" {
		return query, nil
	}

	return bson.D{{Key: field, Value: sorter.sortDirection()}}, nil
}
