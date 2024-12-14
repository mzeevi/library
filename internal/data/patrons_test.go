package data

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

const (
	testTeacherDiscount = 10
	testStudentDiscount = 5
)

var (
	testTeacherCategory = TeacherCategory{CategoryType: Teacher, DiscountPercentage: testTeacherDiscount}
	testStudentCategory = StudentCategory{CategoryType: Student, DiscountPercentage: testStudentDiscount}

	testPatronsIDs []interface{}
	testPatrons    = []interface{}{
		NewPatron("", "John Teacher", "john.teacher@example.com", testTeacherCategory),
		NewPatron("", "Sam Student", "sam.student@example.com", testStudentCategory),
		NewPatron("", "Jane Teacher", "jane.teacher@example.com", testTeacherCategory),
		NewPatron("", "Sandy Student", "sandy.student@example.com", testStudentCategory),
		NewPatron("", "Jim Teacher", "jim.teacher@example.com", testTeacherCategory),
		NewPatron("", "Sue Student", "sue.student@example.com", testStudentCategory),
		NewPatron("", "Jack Teacher", "jack.teacher@example.com", testTeacherCategory),
		NewPatron("", "Steve Student", "steve.student@example.com", testStudentCategory),
		NewPatron("", "Jill Teacher", "jill.teacher@example.com", testTeacherCategory),
		NewPatron("", "Stacy Student", "stacy.student@example.com", testStudentCategory),
		NewPatron("conflict", "Conflict Teacher", "conflict.teacher@example.com", testTeacherCategory),
	}
)

// populatePatronsInDB inserts mockPatrons into the DB.
func (ts *TestSuite) populatePatronsInDB() error {
	client := ts.models.Patrons.Client
	coll := client.Database(ts.models.Patrons.Database).Collection(ts.models.Patrons.Collection)

	res, err := coll.InsertMany(ts.ctx, testPatrons)
	if err != nil {
		return err
	}

	testPatronsIDs = res.InsertedIDs

	return err
}

// deletePatronsFromDB deletes test Patrons from the DB.
func (ts *TestSuite) deletePatronsFromDB(filter PatronFilter) error {
	client := ts.models.Patrons.Client
	coll := client.Database(ts.models.Patrons.Database).Collection(ts.models.Patrons.Collection)

	queryFilter, err := buildPatronFilter(filter)
	if err != nil {
		return err
	}

	_, err = coll.DeleteMany(ts.ctx, queryFilter)

	return err
}

func (ts *TestSuite) TestPatronInsert() {
	t := ts.T()

	tests := []struct {
		name        string
		patron      Patron
		expectError bool
	}{
		{
			name: "ValidPatron",
			patron: Patron{
				Name:     "Test",
				Email:    "test@test.com",
				Category: testTeacherCategory,
			},
			expectError: false,
		},
		{
			name: "PatronWithConflictingID",
			patron: Patron{
				ID:       "conflict",
				Name:     "Conflicting Patron",
				Email:    "conflicting@test.com",
				Category: testTeacherCategory,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			id, err := ts.models.Patrons.Insert(ts.ctx, &tt.patron)
			if tt.expectError {
				assert.ErrorIs(t, err, ErrDuplicateID)
			} else {
				assert.NoError(t, err)
				err = ts.deletePatronsFromDB(PatronFilter{Name: &id})
				assert.NoError(t, err)
			}
		})
	}
}

func (ts *TestSuite) TestGetPatron() {
	t := ts.T()

	tests := []struct {
		name        string
		filter      PatronFilter
		expectedID  string
		expectError bool
	}{
		{
			name:        "FindByID",
			filter:      PatronFilter{ID: ptr(testPatronsIDs[4].(primitive.ObjectID).Hex())},
			expectedID:  testPatronsIDs[4].(primitive.ObjectID).Hex(),
			expectError: false,
		},
		{
			name:        "FindByName",
			filter:      PatronFilter{Name: ptr("John Teacher")},
			expectedID:  testPatronsIDs[0].(primitive.ObjectID).Hex(),
			expectError: false,
		},
		{
			name:        "FindByEmail",
			filter:      PatronFilter{Email: ptr("sam.student@example.com")},
			expectedID:  testPatronsIDs[1].(primitive.ObjectID).Hex(),
			expectError: false,
		},
		{
			name:        "FindByCategory",
			filter:      PatronFilter{Category: testTeacherCategory},
			expectedID:  testPatronsIDs[0].(primitive.ObjectID).Hex(),
			expectError: false,
		},
		{
			name:        "NonExistent",
			filter:      PatronFilter{Name: ptr("Nonexistent patron")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patron, err := ts.models.Patrons.Get(ts.ctx, tt.filter)
			if tt.expectError {
				assert.ErrorIs(t, err, ErrDocumentNotFound)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, patron.ID, tt.expectedID)
			}
		})
	}
}

func (ts *TestSuite) TestGetAllPatrons() {
	t := ts.T()

	tests := []struct {
		name             string
		filter           PatronFilter
		paginator        Paginator
		expectedIDs      []string
		expectedCount    int
		expectedMetadata Metadata
		expectError      bool
	}{
		{
			name:      "FilterByCategoryTeacher",
			filter:    PatronFilter{Category: testTeacherCategory},
			paginator: Paginator{Page: 1, PageSize: 3},
			expectedIDs: []string{
				testPatronsIDs[0].(primitive.ObjectID).Hex(),
				testPatronsIDs[2].(primitive.ObjectID).Hex(),
				testPatronsIDs[4].(primitive.ObjectID).Hex(),
			},
			expectedCount: 3,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     3,
				FirstPage:    1,
				LastPage:     2,
				TotalRecords: 6,
			},
			expectError: false,
		},
		{
			name:      "FilterByCategoryStudent",
			filter:    PatronFilter{Category: testStudentCategory},
			paginator: Paginator{Page: 1, PageSize: 3},
			expectedIDs: []string{
				testPatronsIDs[1].(primitive.ObjectID).Hex(),
				testPatronsIDs[3].(primitive.ObjectID).Hex(),
				testPatronsIDs[5].(primitive.ObjectID).Hex(),
			},
			expectedCount: 3,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     3,
				FirstPage:    1,
				LastPage:     2,
				TotalRecords: 5,
			},
			expectError: false,
		},
		{
			name: "FilterByName",
			filter: PatronFilter{
				Name: ptr("Jane Teacher"),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testPatronsIDs[2].(primitive.ObjectID).Hex(),
			},
			expectedCount: 1,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectError: false,
		},
		{
			name: "FilterByEmail",
			filter: PatronFilter{
				Email: ptr("sam.student@example.com"),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testPatronsIDs[1].(primitive.ObjectID).Hex(),
			},
			expectedCount: 1,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectError: false,
		},
		{
			name:      "PaginationBoundary",
			filter:    PatronFilter{Category: testTeacherCategory},
			paginator: Paginator{Page: 2, PageSize: 2},
			expectedIDs: []string{
				testPatronsIDs[4].(primitive.ObjectID).Hex(),
				testPatronsIDs[6].(primitive.ObjectID).Hex(),
			},
			expectedCount: 2,
			expectedMetadata: Metadata{
				CurrentPage:  2,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     3,
				TotalRecords: 6,
			},
			expectError: false,
		},
		{
			name: "FilterByNonExistingName",
			filter: PatronFilter{
				Name: ptr("Nonexistent Patron"),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByMaxCreatedAt",
			filter: PatronFilter{
				MaxCreatedAt: ptr(time.Now().Add(-24 * time.Hour)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByMinCreatedAt",
			filter: PatronFilter{
				MinCreatedAt: ptr(time.Now().Add(-48 * time.Hour)),
			},
			paginator: Paginator{Page: 3, PageSize: 2},
			expectedIDs: []string{
				testPatronsIDs[4].(primitive.ObjectID).Hex(),
				testPatronsIDs[5].(primitive.ObjectID).Hex(),
			},
			expectedCount: 2,
			expectedMetadata: Metadata{
				CurrentPage:  3,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     6,
				TotalRecords: 11,
			},
			expectError: false,
		},
		{
			name: "FilterByMinUpdatedAt",
			filter: PatronFilter{
				MinUpdatedAt: ptr(time.Now().Add(-12 * time.Hour)),
			},
			paginator:        Paginator{Page: 1, PageSize: 2},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByVersion",
			filter: PatronFilter{
				Version: ptr(int32(0)),
			},
			paginator: Paginator{Page: 1, PageSize: 3},
			expectedIDs: []string{
				testPatronsIDs[0].(primitive.ObjectID).Hex(),
				testPatronsIDs[1].(primitive.ObjectID).Hex(),
				testPatronsIDs[2].(primitive.ObjectID).Hex(),
			},
			expectedCount: 3,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     3,
				FirstPage:    1,
				LastPage:     4,
				TotalRecords: 11,
			},
			expectError: false,
		},
		{
			name: "FilterByDateRange",
			filter: PatronFilter{
				MinCreatedAt: ptr(time.Now().Add(-72 * time.Hour)),
				MaxCreatedAt: ptr(time.Now().Add(-24 * time.Hour)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByCategoryAndEmail",
			filter: PatronFilter{
				Category: testStudentCategory,
				Email:    ptr("sam.student@example.com"),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testPatronsIDs[1].(primitive.ObjectID).Hex(),
			},
			expectedCount: 1,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectError: false,
		},
		{
			name:          "PaginationBeyondAvailableData",
			filter:        PatronFilter{},
			paginator:     Paginator{Page: 1000, PageSize: 10},
			expectedIDs:   []string{},
			expectedCount: 0,
			expectedMetadata: Metadata{
				CurrentPage:  1000,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     2,
				TotalRecords: 11,
			},
			expectError: false,
		},
		{
			name: "FilterByNameAndCategory",
			filter: PatronFilter{
				Name:     ptr("Jim Teacher"),
				Category: testTeacherCategory,
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testPatronsIDs[4].(primitive.ObjectID).Hex(),
			},
			expectedCount: 1,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patrons, metadata, err := ts.models.Patrons.GetAll(ts.ctx, tt.filter, tt.paginator)
			if tt.expectError {
				assert.Error(t, err)
			}

			assert.NoError(t, err)
			assert.Len(t, patrons, tt.expectedCount)

			var actualIDs []string
			for _, patron := range patrons {
				actualIDs = append(actualIDs, patron.ID)
			}
			assert.ElementsMatch(t, tt.expectedIDs, actualIDs)

			assert.Equal(t, tt.expectedMetadata, metadata)
		})
	}
}

func (ts *TestSuite) TestUpdatePatron() {
	t := ts.T()

	tests := []struct {
		name          string
		initialPatron Patron
		updateData    Patron
		expectError   bool
		expected      *Patron
	}{
		{
			name: "UpdateExistingPatron",
			initialPatron: Patron{
				Name:     "Test",
				Email:    "test@test.com",
				Category: testStudentCategory,
			},
			updateData: Patron{
				Name:     "Test",
				Email:    "test2@test.com",
				Category: testStudentCategory,
			},
			expectError: false,
			expected: &Patron{
				Name:     "Test",
				Email:    "test2@test.com",
				Category: testStudentCategory,
			},
		},
		{
			name: "NonExistingPatron",
			initialPatron: Patron{
				ID: "995cb5a4d3ddbde5ebeecc1f",
			},
			updateData: Patron{
				ID:   "995cb5a4d3ddbde5ebeecc1f",
				Name: "Non-existent Patron",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			if tt.expectError {
				err := ts.models.Patrons.Update(ts.ctx, PatronFilter{ID: &tt.initialPatron.ID}, &tt.updateData)
				assert.ErrorIs(t, err, ErrEditConflict)
			} else {
				id, err := ts.models.Patrons.Insert(ts.ctx, &tt.initialPatron)
				assert.NoError(t, err)

				err = ts.models.Patrons.Update(ts.ctx, PatronFilter{ID: &id}, &tt.updateData)
				assert.NoError(t, err)

				updatedPatron, err := ts.models.Patrons.Get(ts.ctx, PatronFilter{ID: &id})
				assert.NoError(t, err)

				assert.Equal(t, updatedPatron.Version, int32(1))

				err = ts.deletePatronsFromDB(PatronFilter{ID: &id})
				assert.NoError(t, err)
			}
		})
	}
}

func (ts *TestSuite) TestDeletePatron() {
	t := ts.T()

	tests := []struct {
		name          string
		initialPatron Patron
		expectError   bool
	}{
		{
			name: "ExistingPatron",
			initialPatron: Patron{
				Name:     "Patron to be deleted",
				Category: testStudentCategory,
			},
			expectError: false,
		},
		{
			name: "NonExistingPatron",
			initialPatron: Patron{
				ID: "995cb5a4d3ddbde5ebeecc1e",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			if tt.expectError {
				err := ts.models.Patrons.Delete(ts.ctx, PatronFilter{ID: &tt.initialPatron.ID})
				assert.ErrorIs(t, err, ErrDocumentNotFound)
			} else {
				id, err := ts.models.Patrons.Insert(ts.ctx, &tt.initialPatron)
				assert.NoError(t, err)

				err = ts.models.Patrons.Delete(ts.ctx, PatronFilter{ID: &id})
				assert.NoError(t, err)

				assert.NoError(t, err)
				_, err = ts.models.Patrons.Get(ts.ctx, PatronFilter{ID: &id})
				assert.ErrorIs(t, err, ErrDocumentNotFound)
			}
		})
	}
}
