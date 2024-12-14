package data

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

var (
	testBooksIDs []interface{}
	testBooks    = []interface{}{
		NewBook("", "Test The Great Adventure", "978-1-23456-789-0", 300, 1, 5, []string{"John Doe"}, []string{"Fiction Press"}, []string{"Adventure", "Fantasy"}, time.Date(2023, time.May, 12, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test A Journey Beyond", "978-0-98765-432-2", 350, 2, 4, []string{"Alice Johnson"}, []string{"Imagination Books"}, []string{"Fantasy", "Adventure"}, time.Date(2022, time.June, 5, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test Science Explained", "978-0-11223-456-7", 400, 1, 3, []string{"Bob Smith"}, []string{"Knowledge Publications"}, []string{"Non-fiction", "Science"}, time.Date(2021, time.September, 18, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test The Mystery of Shadows", "978-1-22334-567-8", 250, 1, 2, []string{"Claire Adams"}, []string{"Mystery House"}, []string{"Mystery", "Thriller"}, time.Date(2020, time.November, 25, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test The Last Sunset", "978-1-23345-678-9", 500, 1, 6, []string{"Michael Young"}, []string{"Sunset Publishing"}, []string{"Romance", "Drama"}, time.Date(2021, time.March, 2, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test Cooking Secrets", "978-0-54321-987-6", 150, 1, 10, []string{"Sarah Lee"}, []string{"Culinary Creations"}, []string{"Cooking", "Lifestyle"}, time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test Tech Innovations", "978-0-89765-123-4", 200, 3, 8, []string{"David Green", "Eva White"}, []string{"TechBooks Publishing"}, []string{"Technology", "Innovation"}, time.Date(2022, time.April, 7, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test Ancient Legends", "978-1-23456-789-1", 300, 2, 7, []string{"Nina Scott"}, []string{"History Publishing"}, []string{"History", "Legends"}, time.Date(2021, time.October, 12, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test In the Depths of the Ocean", "978-0-98765-432-3", 220, 1, 3, []string{"Jack Carter"}, []string{"Oceanic Press"}, []string{"Adventure", "Oceanography"}, time.Date(2022, time.February, 18, 0, 0, 0, 0, time.UTC)),
		NewBook("", "Test Mystic Forest", "978-1-11223-334-5", 380, 2, 5, []string{"Laura Mills"}, []string{"Fiction Publishers"}, []string{"Fantasy", "Adventure"}, time.Date(2020, time.August, 20, 0, 0, 0, 0, time.UTC)),
		NewBook("conflict", "Test Conflict", "978-1-23456-214-0", 0, 1, 1, []string{"Con Doe"}, []string{"Conflict"}, []string{"Conflict"}, time.Date(1999, time.May, 12, 0, 0, 0, 0, time.UTC)),
	}
)

// populateBooksInDB inserts test Books into the DB.
func (ts *TestSuite) populateBooksInDB() error {
	client := ts.models.Books.Client
	coll := client.Database(ts.models.Books.Database).Collection(ts.models.Books.Collection)

	res, err := coll.InsertMany(ts.ctx, testBooks)
	if err != nil {
		return err
	}

	testBooksIDs = res.InsertedIDs

	return nil
}

// deleteBooksFromDB deletes test Books from the DB.
func (ts *TestSuite) deleteBooksFromDB(filter BookFilter) error {
	client := ts.models.Books.Client
	coll := client.Database(ts.models.Books.Database).Collection(ts.models.Books.Collection)

	queryFilter, err := buildBookFilter(filter)
	if err != nil {
		return err
	}

	_, err = coll.DeleteMany(ts.ctx, queryFilter)

	return err
}

func (ts *TestSuite) TestBookInsert() {
	t := ts.T()

	tests := []struct {
		name        string
		book        Book
		expectError bool
	}{
		{
			name: "ValidBook",
			book: Book{
				Title:       "Valid",
				Pages:       10000000,
				Edition:     10000000,
				PublishedAt: time.Date(1900, time.August, 20, 0, 0, 0, 0, time.UTC),
				Authors:     []string{"Author 1"},
				Publishers:  []string{"Publisher 1"},
				Genres:      []string{"Fiction"},
			},
			expectError: false,
		},
		{
			name: "BookWithConflictingID",
			book: Book{
				ID:          "conflict",
				Title:       "Conflicting Book",
				Pages:       100,
				Edition:     1,
				PublishedAt: time.Now(),
				Authors:     []string{"Author 3"},
				Publishers:  []string{"Publisher 3"},
				Genres:      []string{"Fantasy"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			id, err := ts.models.Books.Insert(ts.ctx, &tt.book)
			if tt.expectError {
				assert.ErrorIs(t, err, ErrDuplicateID)
			} else {
				assert.NoError(t, err)
				err = ts.deleteBooksFromDB(BookFilter{Title: &id})
				assert.NoError(t, err)
			}
		})
	}
}

func (ts *TestSuite) TestGetBook() {
	t := ts.T()

	tests := []struct {
		name        string
		filter      BookFilter
		expectedID  string
		expectError bool
	}{
		{
			name:        "FindByID",
			filter:      BookFilter{ID: ptr(testBooksIDs[4].(primitive.ObjectID).Hex())},
			expectedID:  testBooksIDs[4].(primitive.ObjectID).Hex(),
			expectError: false,
		},
		{
			name:        "FindByTitle",
			filter:      BookFilter{Title: ptr("The Great Adventure")},
			expectedID:  testBooksIDs[0].(primitive.ObjectID).Hex(),
			expectError: false,
		},
		{
			name:        "FindByAuthor",
			filter:      BookFilter{Authors: []string{"John Doe"}},
			expectedID:  testBooksIDs[0].(primitive.ObjectID).Hex(),
			expectError: false,
		},
		{
			name:        "FindByPublisher",
			filter:      BookFilter{Publishers: []string{"Fiction Press"}},
			expectedID:  testBooksIDs[0].(primitive.ObjectID).Hex(),
			expectError: false,
		},
		{
			name:        "NonExistent",
			filter:      BookFilter{Title: ptr("Nonexistent book")},
			expectedID:  testBooksIDs[0].(primitive.ObjectID).Hex(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book, err := ts.models.Books.Get(ts.ctx, tt.filter)
			if tt.expectError {
				assert.ErrorIs(t, err, ErrDocumentNotFound)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, book.ID, tt.expectedID)
			}
		})
	}
}

func (ts *TestSuite) TestGetAllBooks() {
	t := ts.T()

	tests := []struct {
		name             string
		filter           BookFilter
		paginator        Paginator
		expectedIDs      []string
		expectedCount    int
		expectedMetadata Metadata
		expectError      bool
	}{
		{
			name: "FilterByMinPages",
			filter: BookFilter{
				Title:    ptr("Test"),
				MinPages: ptr(300),
			},
			paginator: Paginator{Page: 1, PageSize: 1},
			expectedIDs: []string{
				testBooksIDs[0].(primitive.ObjectID).Hex(),
			},
			expectedCount: 1,
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     1,
				FirstPage:    1,
				LastPage:     6,
				TotalRecords: 6,
			},
			expectError: false,
		},
		{
			name: "FilterByMaxPages",
			filter: BookFilter{
				Title:    ptr("Test"),
				MaxPages: ptr(250),
			},
			paginator: Paginator{Page: 1, PageSize: 2},
			expectedIDs: []string{
				testBooksIDs[3].(primitive.ObjectID).Hex(),
				testBooksIDs[5].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     3,
				TotalRecords: 5,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "FilterByPagesRange",
			filter: BookFilter{
				Title:    ptr("Test"),
				MinPages: ptr(200),
				MaxPages: ptr(400),
			},
			paginator: Paginator{Page: 2, PageSize: 4},
			expectedIDs: []string{
				testBooksIDs[6].(primitive.ObjectID).Hex(),
				testBooksIDs[7].(primitive.ObjectID).Hex(),
				testBooksIDs[8].(primitive.ObjectID).Hex(),
				testBooksIDs[9].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  2,
				PageSize:     4,
				FirstPage:    1,
				LastPage:     2,
				TotalRecords: 8,
			},
			expectedCount: 4,
			expectError:   false,
		},
		{
			name: "FilterByPublishedAfter",
			filter: BookFilter{
				Title:          ptr("Test"),
				MinPublishedAt: ptr(time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testBooksIDs[0].(primitive.ObjectID).Hex(),
				testBooksIDs[1].(primitive.ObjectID).Hex(),
				testBooksIDs[5].(primitive.ObjectID).Hex(),
				testBooksIDs[6].(primitive.ObjectID).Hex(),
				testBooksIDs[8].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 5,
			},
			expectedCount: 5,
			expectError:   false,
		},
		{
			name: "FilterByPublishedBefore",
			filter: BookFilter{
				Title:          ptr("Test"),
				MaxPublishedAt: ptr(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator: Paginator{Page: 1, PageSize: 2},
			expectedIDs: []string{
				testBooksIDs[3].(primitive.ObjectID).Hex(),
				testBooksIDs[9].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     2,
				TotalRecords: 3,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "FilterByPublishDateRange",
			filter: BookFilter{
				Title:          ptr("Test"),
				MinPublishedAt: ptr(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
				MaxPublishedAt: ptr(time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testBooksIDs[1].(primitive.ObjectID).Hex(),
				testBooksIDs[2].(primitive.ObjectID).Hex(),
				testBooksIDs[4].(primitive.ObjectID).Hex(),
				testBooksIDs[6].(primitive.ObjectID).Hex(),
				testBooksIDs[7].(primitive.ObjectID).Hex(),
				testBooksIDs[8].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 6,
			},
			expectedCount: 6,
			expectError:   false,
		},
		{
			name: "FilterByGenresAndMinPages",
			filter: BookFilter{
				Title:    ptr("Test"),
				Genres:   []string{"Adventure"},
				MinPages: ptr(300),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testBooksIDs[0].(primitive.ObjectID).Hex(),
				testBooksIDs[1].(primitive.ObjectID).Hex(),
				testBooksIDs[9].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 3,
			},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "FilterByAuthorsAndDateRange",
			filter: BookFilter{
				Title:          ptr("Test"),
				Authors:        []string{"Nina Scott"},
				MinPublishedAt: ptr(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
				MaxPublishedAt: ptr(time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testBooksIDs[7].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "FilterByMinCreatedAt",
			filter: BookFilter{
				Title:        ptr("Test"),
				MinCreatedAt: ptr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator: Paginator{Page: 1, PageSize: 2},
			expectedIDs: []string{
				testBooksIDs[0].(primitive.ObjectID).Hex(),
				testBooksIDs[1].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     6,
				TotalRecords: 11,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "FilterByMaxCreatedAt",
			filter: BookFilter{
				Title:        ptr("Test"),
				MaxCreatedAt: ptr(time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByMinAndMaxCreatedAt",
			filter: BookFilter{
				Title:        ptr("Test"),
				MinCreatedAt: ptr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				MaxCreatedAt: ptr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedMetadata: Metadata{},
			expectedCount:    0,
			expectError:      false,
		},
		{
			name: "FilterByMinUpdatedAt",
			filter: BookFilter{
				Title:        ptr("Test"),
				MinUpdatedAt: ptr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator: Paginator{Page: 2, PageSize: 2},
			expectedIDs: []string{
				testBooksIDs[2].(primitive.ObjectID).Hex(),
				testBooksIDs[3].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  2,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     6,
				TotalRecords: 11,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "FilterByMaxUpdatedAt",
			filter: BookFilter{
				Title:        ptr("Test"),
				MaxUpdatedAt: ptr(time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedCount:    0,
			expectedMetadata: Metadata{},
			expectError:      false,
		},
		{
			name: "FilterByExactVersion",
			filter: BookFilter{
				Title:   ptr("Test"),
				Version: ptr(int32(0)),
			},
			paginator: Paginator{Page: 2, PageSize: 5},
			expectedIDs: []string{
				testBooksIDs[5].(primitive.ObjectID).Hex(),
				testBooksIDs[6].(primitive.ObjectID).Hex(),
				testBooksIDs[7].(primitive.ObjectID).Hex(),
				testBooksIDs[8].(primitive.ObjectID).Hex(),
				testBooksIDs[9].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  2,
				PageSize:     5,
				FirstPage:    1,
				LastPage:     3,
				TotalRecords: 11,
			},
			expectedCount: 5,
			expectError:   false,
		},
		{
			name: "FilterByMinCreatedAtAndVersion",
			filter: BookFilter{
				Title:        ptr("Test"),
				MinCreatedAt: ptr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				Version:      ptr(int32(2)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedMetadata: Metadata{},
			expectedCount:    0,
			expectError:      false,
		},
		{
			name: "BoundaryPages",
			filter: BookFilter{
				Title:    ptr("Test"),
				MinPages: ptr(150),
				MaxPages: ptr(150),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testBooksIDs[5].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "AuthorGenreDateFilter",
			filter: BookFilter{
				Title:          ptr("Test"),
				Authors:        []string{"David Green"},
				Genres:         []string{"Technology"},
				MinPublishedAt: ptr(time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)),
				MaxPublishedAt: ptr(time.Date(2022, time.December, 31, 23, 59, 59, 0, time.UTC)),
			},
			paginator: Paginator{Page: 1, PageSize: 10},
			expectedIDs: []string{
				testBooksIDs[6].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 1,
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "PaginationExceedsData",
			filter: BookFilter{
				Title:    ptr("Test"),
				MinPages: ptr(300),
			},
			paginator:   Paginator{Page: 100, PageSize: 10},
			expectedIDs: []string{},
			expectedMetadata: Metadata{
				CurrentPage:  100,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 6,
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "NoMatchingCreatedAt",
			filter: BookFilter{
				Title:        ptr("Test"),
				MinCreatedAt: ptr(time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedMetadata: Metadata{},
			expectedCount:    0,
			expectError:      false,
		},
		{
			name: "NoMatchingVersion",
			filter: BookFilter{
				Title:   ptr("Test"),
				Version: ptr(int32(999)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedMetadata: Metadata{},
			expectedCount:    0,
			expectError:      false,
		},
		{
			name: "NoMatchingPagesRange",
			filter: BookFilter{
				Title:    ptr("Test"),
				MinPages: ptr(1000),
				MaxPages: ptr(2000),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedMetadata: Metadata{},
			expectedCount:    0,
			expectError:      false,
		},
		{
			name: "NoMatchingPublishDate",
			filter: BookFilter{
				Title:          ptr("Test"),
				MinPublishedAt: ptr(time.Date(2030, time.January, 1, 0, 0, 0, 0, time.UTC)),
				MaxPublishedAt: ptr(time.Date(2031, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
			paginator:        Paginator{Page: 1, PageSize: 10},
			expectedIDs:      []string{},
			expectedMetadata: Metadata{},
			expectedCount:    0,
			expectError:      false,
		},
		{
			name: "FilterByMinCopies",
			filter: BookFilter{
				Title:     ptr("Test"),
				MinCopies: ptr(5),
			},
			paginator: Paginator{Page: 1, PageSize: 2},
			expectedIDs: []string{
				testBooksIDs[0].(primitive.ObjectID).Hex(),
				testBooksIDs[4].(primitive.ObjectID).Hex(),
			},
			expectedMetadata: Metadata{
				CurrentPage:  1,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     3,
				TotalRecords: 6,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "FilterByBorrowedCopiesRange",
			filter: BookFilter{
				Title:             ptr("Test"),
				MinBorrowedCopies: ptr(2),
				MaxBorrowedCopies: ptr(4),
			},
			paginator:        Paginator{Page: 1, PageSize: 3},
			expectedIDs:      []string{},
			expectedMetadata: Metadata{},
			expectedCount:    0,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			books, metadata, err := ts.models.Books.GetAll(ts.ctx, tt.filter, tt.paginator)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, books, tt.expectedCount)

				var actualIDs []string
				for _, book := range books {
					actualIDs = append(actualIDs, book.ID)
				}
				assert.ElementsMatch(t, tt.expectedIDs, actualIDs)

				assert.Equal(t, tt.expectedMetadata, metadata)
			}
		})
	}
}

func (ts *TestSuite) TestUpdateBook() {
	t := ts.T()

	tests := []struct {
		name        string
		initialBook Book
		updateData  Book
		expectError bool
		expected    *Book
	}{
		{
			name: "UpdateExistingBook",
			initialBook: Book{
				Title:  "Test",
				Pages:  100,
				Genres: []string{"Fiction"},
			},
			updateData: Book{
				Title:  "Test",
				Pages:  200,
				Genres: []string{"Fiction", "Adventure"},
			},
			expectError: false,
			expected: &Book{
				Title:  "Test",
				Pages:  200,
				Genres: []string{"Fiction", "Adventure"},
			},
		},
		{
			name: "NonExistingBook",
			initialBook: Book{
				ID: "995cb5a4d3ddbde5ebeecc1f",
			},
			updateData: Book{
				ID:    "995cb5a4d3ddbde5ebeecc1f",
				Title: "Non-existent Book",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			if tt.expectError {
				err := ts.models.Books.Update(ts.ctx, BookFilter{ID: &tt.initialBook.ID}, &tt.updateData)
				assert.ErrorIs(t, err, ErrEditConflict)
			} else {
				id, err := ts.models.Books.Insert(ts.ctx, &tt.initialBook)
				assert.NoError(t, err)

				err = ts.models.Books.Update(ts.ctx, BookFilter{ID: &id}, &tt.updateData)
				assert.NoError(t, err)

				updatedBook, err := ts.models.Books.Get(ts.ctx, BookFilter{ID: &id})
				assert.NoError(t, err)

				assert.Equal(t, updatedBook.Version, int32(1))

				err = ts.deleteBooksFromDB(BookFilter{ID: &id})
				assert.NoError(t, err)
			}
		})
	}
}

func (ts *TestSuite) TestDeleteBook() {
	t := ts.T()

	tests := []struct {
		name        string
		initialBook Book
		expectError bool
	}{
		{
			name: "ExistingBook",
			initialBook: Book{
				Title: "Book to be deleted",
			},
			expectError: false,
		},
		{
			name: "NonExistingBook",
			initialBook: Book{
				ID: "995cb5a4d3ddbde5ebeecc1e",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			if tt.expectError {
				err := ts.models.Books.Delete(ts.ctx, BookFilter{ID: &tt.initialBook.ID})
				assert.ErrorIs(t, err, ErrDocumentNotFound)
			} else {
				id, err := ts.models.Books.Insert(ts.ctx, &tt.initialBook)
				assert.NoError(t, err)

				err = ts.models.Books.Delete(ts.ctx, BookFilter{ID: &id})
				assert.NoError(t, err)

				_, err = ts.models.Books.Get(ts.ctx, BookFilter{ID: &id})
				assert.ErrorIs(t, err, ErrDocumentNotFound)
			}
		})
	}
}
