package data

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestMarkBookAsBorrowed(t *testing.T) {
	tests := []struct {
		name                  string
		book                  *Book
		expectedError         error
		initialBorrowedStatus bool
	}{
		{
			name:                  "AlreadyBorrowedBook",
			book:                  &Book{Title: "Go Programming", Borrowed: true},
			expectedError:         errors.New(errBookAlreadyBorrowed),
			initialBorrowedStatus: true,
		},
		{
			name:                  "SuccessfullyBorrowBook",
			book:                  &Book{Title: "Go Programming", Borrowed: false},
			expectedError:         nil,
			initialBorrowedStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.book.Borrowed != tt.initialBorrowedStatus {
				t.Errorf("Expected initial Borrowed status to be %v, but got %v", tt.initialBorrowedStatus, tt.book.Borrowed)
			}

			err := tt.book.markBookAsBorrowed()

			if err != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			if err != nil {
				if tt.book.Borrowed != tt.initialBorrowedStatus {
					t.Errorf("Expected book Borrowed status to remain %v, but it was changed to %v", tt.initialBorrowedStatus, tt.book.Borrowed)
				}
			} else {
				if !tt.book.Borrowed {
					t.Errorf("Expected book to be marked as borrowed, but it was not")
				}
			}
		})
	}
}

func TestMarkBookAsNotBorrowed(t *testing.T) {
	type args struct {
		book *Book
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "mark book as not borrowed",
			args: args{
				book: &Book{
					Title:    "Test Book",
					ISBN:     "1234567890",
					Authors:  []string{"Author One"},
					Borrowed: true,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.book.markBookAsNotBorrowed()
			if tt.args.book.Borrowed != tt.want {
				t.Errorf("markBookAsNotBorrowed() = %v, want %v", tt.args.book.Borrowed, tt.want)
			}
		})
	}
}

func TestNewBook(t *testing.T) {
	now := time.Now()
	type args struct {
		title          string
		isbn           string
		authors        []string
		publishers     []string
		genres         []string
		pages          int
		edition        int
		published      time.Time
		borrowDuration time.Duration
	}
	tests := []struct {
		name string
		args args
		want *Book
	}{
		{
			name: "CreateSuccessfully",
			args: args{
				title:          "Test Title",
				isbn:           "1234567890",
				authors:        []string{"Author1", "Author2"},
				publishers:     []string{"Publisher1"},
				genres:         []string{"Fiction", "Adventure"},
				pages:          350,
				edition:        1,
				published:      now,
				borrowDuration: 14 * 24 * time.Hour,
			},
			want: &Book{
				Title:          "Test Title",
				ISBN:           "1234567890",
				Authors:        []string{"Author1", "Author2"},
				Publishers:     []string{"Publisher1"},
				Genres:         []string{"Fiction", "Adventure"},
				Pages:          350,
				Edition:        1,
				Published:      now,
				BorrowDuration: 14 * 24 * time.Hour,
				Borrowed:       false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBook(tt.args.title, tt.args.isbn, tt.args.authors, tt.args.publishers, tt.args.genres, tt.args.pages, tt.args.edition, tt.args.published, tt.args.borrowDuration)
			if got.Title != tt.want.Title ||
				got.ISBN != tt.want.ISBN ||
				!reflect.DeepEqual(got.Authors, tt.want.Authors) ||
				!reflect.DeepEqual(got.Publishers, tt.want.Publishers) ||
				!reflect.DeepEqual(got.Genres, tt.want.Genres) ||
				got.Pages != tt.want.Pages ||
				got.Edition != tt.want.Edition ||
				!got.Published.Equal(tt.want.Published) ||
				got.BorrowDuration != tt.want.BorrowDuration ||
				got.Borrowed != tt.want.Borrowed {
				t.Errorf("NewBook() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUpdateBook(t *testing.T) {
	now := time.Now()
	newTime := now.AddDate(0, 0, 1)
	newDuration := 21 * 24 * time.Hour

	type args struct {
		title          *string
		isbn           *string
		authors        *[]string
		publishers     *[]string
		genres         *[]string
		pages          *int
		edition        *int
		published      *time.Time
		borrowDuration *time.Duration
	}
	tests := []struct {
		name string
		book *Book
		args args
		want *Book
	}{
		{
			name: "update title and authors",
			book: &Book{
				Title:   "Old Title",
				Authors: []string{"Old Author"},
			},
			args: args{
				title:   ptrStr("New Title"),
				authors: &[]string{"New Author1", "New Author2"},
			},
			want: &Book{
				Title:   "New Title",
				Authors: []string{"New Author1", "New Author2"},
			},
		},
		{
			name: "update ISBN, pages, and edition",
			book: &Book{
				ISBN:    "1234567890",
				Pages:   300,
				Edition: 1,
			},
			args: args{
				isbn:    ptrStr("0987654321"),
				pages:   ptrInt(400),
				edition: ptrInt(2),
			},
			want: &Book{
				ISBN:    "0987654321",
				Pages:   400,
				Edition: 2,
			},
		},
		{
			name: "update publishers and genres",
			book: &Book{
				Publishers: []string{"Old Publisher"},
				Genres:     []string{"Fiction"},
			},
			args: args{
				publishers: &[]string{"New Publisher1", "New Publisher2"},
				genres:     &[]string{"Adventure", "Mystery"},
			},
			want: &Book{
				Publishers: []string{"New Publisher1", "New Publisher2"},
				Genres:     []string{"Adventure", "Mystery"},
			},
		},
		{
			name: "update published date and borrow duration",
			book: &Book{
				Published:      now,
				BorrowDuration: 14 * 24 * time.Hour,
			},
			args: args{
				published:      &newTime,
				borrowDuration: &newDuration,
			},
			want: &Book{
				Published:      newTime,
				BorrowDuration: newDuration,
			},
		},
		{
			name: "update all fields",
			book: &Book{
				Title:          "Old Title",
				ISBN:           "1234567890",
				Authors:        []string{"Author1"},
				Publishers:     []string{"Publisher1"},
				Genres:         []string{"Fiction"},
				Pages:          300,
				Edition:        1,
				Published:      now,
				BorrowDuration: 14 * 24 * time.Hour,
			},
			args: args{
				title:          ptrStr("New Title"),
				isbn:           ptrStr("0987654321"),
				authors:        &[]string{"Author1", "Author2"},
				publishers:     &[]string{"Publisher2"},
				genres:         &[]string{"Adventure"},
				pages:          ptrInt(400),
				edition:        ptrInt(2),
				published:      &newTime,
				borrowDuration: &newDuration,
			},
			want: &Book{
				Title:          "New Title",
				ISBN:           "0987654321",
				Authors:        []string{"Author1", "Author2"},
				Publishers:     []string{"Publisher2"},
				Genres:         []string{"Adventure"},
				Pages:          400,
				Edition:        2,
				Published:      newTime,
				BorrowDuration: newDuration,
			},
		},
		{
			name: "no updates",
			book: &Book{
				Title:          "No Update",
				ISBN:           "1111111111",
				Authors:        []string{"Author1"},
				Publishers:     []string{"Publisher1"},
				Genres:         []string{"Fiction"},
				Pages:          100,
				Edition:        1,
				Published:      now,
				BorrowDuration: 7 * 24 * time.Hour,
			},
			args: args{},
			want: &Book{
				Title:          "No Update",
				ISBN:           "1111111111",
				Authors:        []string{"Author1"},
				Publishers:     []string{"Publisher1"},
				Genres:         []string{"Fiction"},
				Pages:          100,
				Edition:        1,
				Published:      now,
				BorrowDuration: 7 * 24 * time.Hour,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.book.UpdateBook(tt.args.title, tt.args.isbn, tt.args.authors, tt.args.publishers, tt.args.genres, tt.args.pages, tt.args.edition, tt.args.published, tt.args.borrowDuration)
			if !reflect.DeepEqual(tt.book, tt.want) {
				t.Errorf("UpdateBook() = %+v, want %+v", tt.book, tt.want)
			}
		})
	}
}

func TestSearchBooks(t *testing.T) {
	books := []*Book{
		{
			ID:         1,
			Title:      "The Great Gatsby",
			ISBN:       "1234567890",
			Authors:    []string{"F. Scott Fitzgerald"},
			Publishers: []string{"Scribner"},
			Genres:     []string{"Fiction"},
			Pages:      218,
			Edition:    1,
			Published:  time.Date(1925, 4, 10, 0, 0, 0, 0, time.UTC),
			Borrowed:   false,
		},
		{
			ID:         2,
			Title:      "1984",
			ISBN:       "0987654321",
			Authors:    []string{"George Orwell"},
			Publishers: []string{"Secker & Warburg"},
			Genres:     []string{"Dystopian", "Science Fiction"},
			Pages:      328,
			Edition:    1,
			Published:  time.Date(1949, 6, 8, 0, 0, 0, 0, time.UTC),
			Borrowed:   false,
		},
		{
			ID:         3,
			Title:      "Moby Dick",
			ISBN:       "1112131415",
			Authors:    []string{"Herman Melville"},
			Publishers: []string{"Harper & Brothers"},
			Genres:     []string{"Adventure", "Fiction"},
			Pages:      635,
			Edition:    1,
			Published:  time.Date(1851, 10, 18, 0, 0, 0, 0, time.UTC),
			Borrowed:   false,
		},
		{
			ID:         4,
			Title:      "To Kill a Mockingbird",
			ISBN:       "5556677788",
			Authors:    []string{"Harper Lee"},
			Publishers: []string{"J.B. Lippincott & Co."},
			Genres:     []string{"Fiction"},
			Pages:      281,
			Edition:    1,
			Published:  time.Date(1960, 7, 11, 0, 0, 0, 0, time.UTC),
			Borrowed:   false,
		},
		{
			ID:         5,
			Title:      "Pride and Prejudice",
			ISBN:       "2223334445",
			Authors:    []string{"Jane Austen"},
			Publishers: []string{"T. Egerton"},
			Genres:     []string{"Romance", "Fiction"},
			Pages:      432,
			Edition:    1,
			Published:  time.Date(1813, 1, 28, 0, 0, 0, 0, time.UTC),
			Borrowed:   false,
		},
	}

	tests := []struct {
		name     string
		criteria SearchCriteria
		want     []*Book
	}{
		{
			name: "ByTileExactMatch",
			criteria: SearchCriteria{
				Title: ptrStr("The Great Gatsby"),
			},
			want: []*Book{books[0]},
		},
		{
			name: "TitlePartialMatch",
			criteria: SearchCriteria{
				Title: ptrStr("Pride"),
			},
			want: []*Book{books[4]},
		},
		{
			name: "ByISBN",
			criteria: SearchCriteria{
				ISBN: ptrStr("1112131415"),
			},
			want: []*Book{books[2]},
		},
		{
			name: "ByAuthor",
			criteria: SearchCriteria{
				Authors: &[]string{"Jane Austen"},
			},
			want: []*Book{books[4]},
		},
		{
			name: "ByPublisher",
			criteria: SearchCriteria{
				Publishers: &[]string{"Scribner"},
			},
			want: []*Book{books[0]},
		},
		{
			name: "ByGenre",
			criteria: SearchCriteria{
				Genres: &[]string{"Fiction"},
			},
			want: []*Book{books[0], books[2], books[3], books[4]},
		},
		{
			name: "ByBorrowedStatus",
			criteria: SearchCriteria{
				Borrowed: ptrBool(false),
			},
			want: books,
		},
		{
			name: "ByPageRange",
			criteria: SearchCriteria{
				MinPages: ptrInt(200),
				MaxPages: ptrInt(400),
			},
			want: []*Book{books[0], books[1], books[3]},
		},
		{
			name: "ByEditionRange",
			criteria: SearchCriteria{
				MinEdition: ptrInt(1),
				MaxEdition: ptrInt(1),
			},
			want: books,
		},
		{
			name: "ByPublishedDateRange",
			criteria: SearchCriteria{
				MinPublished: ptrTime(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)),
				MaxPublished: ptrTime(time.Date(1950, 12, 31, 0, 0, 0, 0, time.UTC)),
			},
			want: []*Book{books[0], books[1]},
		},
		{
			name:     "NoCriteria",
			criteria: SearchCriteria{},
			want:     books,
		},
		{
			name: "NoMatchInvalidISBN",
			criteria: SearchCriteria{
				ISBN: ptrStr("0000000000"),
			},
			want: []*Book{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SearchBooks(books, tt.criteria)
			if len(got) != len(tt.want) {
				t.Errorf("SearchBooks() = %v, want %v", got, tt.want)
				return
			}

			for i, book := range got {
				if book.Title != tt.want[i].Title {
					t.Errorf("SearchBooks() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestCheckBorrowed(t *testing.T) {
	type args struct {
		bookBorrowed bool
		borrowed     *bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "MatchFound",
			args: args{
				bookBorrowed: true,
				borrowed:     ptrBool(true),
			},
			want: true,
		},
		{
			name: "NoMatch",
			args: args{
				bookBorrowed: false,
				borrowed:     ptrBool(true),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkBorrowed(tt.args.bookBorrowed, tt.args.borrowed); got != tt.want {
				t.Errorf("checkBorrowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckEdition(t *testing.T) {
	type args struct {
		bookEdition int
		minEdition  *int
		maxEdition  *int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "InRange",
			args: args{
				bookEdition: 2,
				minEdition:  ptrInt(1),
				maxEdition:  ptrInt(3),
			},
			want: true,
		},
		{
			name: "BelowRange",
			args: args{
				bookEdition: 0,
				minEdition:  ptrInt(1),
				maxEdition:  ptrInt(3),
			},
			want: false,
		},
		{
			name: "AboveRange",
			args: args{
				bookEdition: 4,
				minEdition:  ptrInt(1),
				maxEdition:  ptrInt(3),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkEdition(tt.args.bookEdition, tt.args.minEdition, tt.args.maxEdition); got != tt.want {
				t.Errorf("checkEdition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckISBN(t *testing.T) {
	type args struct {
		bookISBN string
		isbn     *string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ExactMatch",
			args: args{
				bookISBN: "1234567890",
				isbn:     ptrStr("1234567890"),
			},
			want: true,
		},
		{
			name: "Mismatch",
			args: args{
				bookISBN: "1234567890",
				isbn:     ptrStr("0987654321"),
			},
			want: false,
		},
		{
			name: "EmptyISBN",
			args: args{
				bookISBN: "",
				isbn:     ptrStr("1234567890"),
			},
			want: false,
		},
		{
			name: "EmptySearchISBN",
			args: args{
				bookISBN: "1234567890",
				isbn:     ptrStr(""),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkISBN(tt.args.bookISBN, tt.args.isbn); got != tt.want {
				t.Errorf("checkISBN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPages(t *testing.T) {
	type args struct {
		bookPages int
		minPages  *int
		maxPages  *int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "InRange",
			args: args{
				bookPages: 200,
				minPages:  ptrInt(100),
				maxPages:  ptrInt(300),
			},
			want: true,
		},
		{
			name: "BelowRange",
			args: args{
				bookPages: 50,
				minPages:  ptrInt(100),
				maxPages:  ptrInt(300),
			},
			want: false,
		},
		{
			name: "AboveRange",
			args: args{
				bookPages: 350,
				minPages:  ptrInt(100),
				maxPages:  ptrInt(300),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPages(tt.args.bookPages, tt.args.minPages, tt.args.maxPages); got != tt.want {
				t.Errorf("checkPages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPublished(t *testing.T) {
	type args struct {
		bookPublished time.Time
		minPublished  *time.Time
		maxPublished  *time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "InRange",
			args: args{
				bookPublished: time.Date(2020, 5, 1, 0, 0, 0, 0, time.UTC),
				minPublished:  ptrTime(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)),
				maxPublished:  ptrTime(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			want: true,
		},
		{
			name: "BeforeRange",
			args: args{
				bookPublished: time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC),
				minPublished:  ptrTime(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)),
				maxPublished:  ptrTime(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			want: false,
		},
		{
			name: "AfterRange",
			args: args{
				bookPublished: time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC),
				minPublished:  ptrTime(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)),
				maxPublished:  ptrTime(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPublished(tt.args.bookPublished, tt.args.minPublished, tt.args.maxPublished); got != tt.want {
				t.Errorf("checkPublished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckStringSlice(t *testing.T) {
	type args struct {
		s             []string
		criteriaSlice *[]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "MatchFound",
			args: args{
				s:             []string{"author1", "author2", "author3"},
				criteriaSlice: ptrStrSlice([]string{"author2"}),
			},
			want: true,
		},
		{
			name: "NoMatch",
			args: args{
				s:             []string{"author1", "author3"},
				criteriaSlice: ptrStrSlice([]string{"author2"}),
			},
			want: false,
		},
		{
			name: "MultipleMatches",
			args: args{
				s:             []string{"author1", "author2", "author3"},
				criteriaSlice: ptrStrSlice([]string{"author2", "author1"}),
			},
			want: true,
		},
		{
			name: "EmptyBookSlice",
			args: args{
				s:             []string{},
				criteriaSlice: ptrStrSlice([]string{"author1"}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkStringSlice(tt.args.s, tt.args.criteriaSlice); got != tt.want {
				t.Errorf("checkStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckTitle(t *testing.T) {
	type args struct {
		bookTitle string
		title     *string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ExactMatch",
			args: args{
				bookTitle: "The Great Gatsby",
				title:     ptrStr("The Great Gatsby"),
			},
			want: true,
		},
		{
			name: "NoMatch",
			args: args{
				bookTitle: "The Great Gatsby",
				title:     ptrStr("Moby Dick"),
			},
			want: false,
		},
		{
			name: "EmptyStringInBookTitle",
			args: args{
				bookTitle: "",
				title:     ptrStr("Great"),
			},
			want: false,
		},
		{
			name: "PartialMatch",
			args: args{
				bookTitle: "The Great Gatsby",
				title:     ptrStr("Gatsby"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkTitle(tt.args.bookTitle, tt.args.title); got != tt.want {
				t.Errorf("checkTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	type args struct {
		slice []string
		str   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "string is in slice",
			args: args{
				slice: []string{"apple", "banana", "cherry"},
				str:   "banana",
			},
			want: true,
		},
		{
			name: "string is not in slice",
			args: args{
				slice: []string{"apple", "banana", "cherry"},
				str:   "orange",
			},
			want: false,
		},
		{
			name: "empty slice",
			args: args{
				slice: []string{},
				str:   "banana",
			},
			want: false,
		},
		{
			name: "empty string",
			args: args{
				slice: []string{"apple", "banana", "cherry"},
				str:   "",
			},
			want: false,
		},
		{
			name: "string is the first element",
			args: args{
				slice: []string{"apple", "banana", "cherry"},
				str:   "apple",
			},
			want: true,
		},
		{
			name: "string is the last element",
			args: args{
				slice: []string{"apple", "banana", "cherry"},
				str:   "cherry",
			},
			want: true,
		},
		{
			name: "string is in slice with repeated elements",
			args: args{
				slice: []string{"apple", "banana", "banana", "cherry"},
				str:   "banana",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.slice, tt.args.str); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBookByTitle(t *testing.T) {
	type args struct {
		title string
		books []*Book
	}
	tests := []struct {
		name    string
		args    args
		want    *Book
		wantErr error
	}{
		{
			name: "FoundBook",
			args: args{
				title: "Test Book",
				books: []*Book{
					{Title: "Test Book", ISBN: "1234567890"},
					{Title: "Another Book", ISBN: "0987654321"},
				},
			},
			want:    &Book{Title: "Test Book", ISBN: "1234567890"},
			wantErr: nil,
		},
		{
			name: "NotFoundBook",
			args: args{
				title: "Nonexistent Book",
				books: []*Book{
					{Title: "Test Book", ISBN: "1234567890"},
					{Title: "Another Book", ISBN: "0987654321"},
				},
			},
			want:    nil,
			wantErr: errors.New(errNonexistentBook),
		},
		{
			name: "EmptyBookSlice",
			args: args{
				title: "Test Book",
				books: []*Book{},
			},
			want:    nil,
			wantErr: errors.New(errNonexistentBook),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBookByTitle(tt.args.title, tt.args.books)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBookByTitle() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("getBookByTitle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchesAllCriteria(t *testing.T) {
	type args struct {
		book     *Book
		criteria SearchCriteria
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "AllCriteriaMatch",
			args: args{
				book: &Book{
					Title:      "The Great Gatsby",
					ISBN:       "1234567890",
					Authors:    []string{"F. Scott Fitzgerald"},
					Publishers: []string{"Scribner"},
					Genres:     []string{"Fiction"},
					Borrowed:   false,
					Pages:      218,
					Edition:    1,
					Published:  time.Date(1925, time.April, 10, 0, 0, 0, 0, time.UTC),
				},
				criteria: SearchCriteria{
					Title:        ptrStr("The Great Gatsby"),
					ISBN:         ptrStr("1234567890"),
					Authors:      &[]string{"F. Scott Fitzgerald"},
					Publishers:   &[]string{"Scribner"},
					Genres:       &[]string{"Fiction"},
					Borrowed:     ptrBool(false),
					MinPages:     ptrInt(200),
					MaxPages:     ptrInt(300),
					MinEdition:   ptrInt(1),
					MaxEdition:   ptrInt(1),
					MinPublished: ptrTime(time.Date(1920, time.January, 1, 0, 0, 0, 0, time.UTC)),
					MaxPublished: ptrTime(time.Date(1930, time.January, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
			want: true,
		},
		{
			name: "NoSpecifiedCriteria",
			args: args{
				book: &Book{
					Title: "1984",
					ISBN:  "0987654321",
					Authors: []string{
						"George Orwell",
					},
					Publishers: []string{"Secker & Warburg"},
					Genres:     []string{"Dystopian", "Science Fiction"},
					Pages:      328,
					Edition:    1,
					Published:  time.Date(1949, time.June, 8, 0, 0, 0, 0, time.UTC),
				},
				criteria: SearchCriteria{},
			},
			want: true,
		},
		{
			name: "AuthorsDoNotMatch",
			args: args{
				book: &Book{
					Title:      "Moby Dick",
					ISBN:       "1112131415",
					Authors:    []string{"Herman Melville"},
					Publishers: []string{"Harper & Brothers"},
					Genres:     []string{"Adventure", "Fiction"},
					Pages:      635,
					Edition:    1,
					Published:  time.Date(1851, time.October, 18, 0, 0, 0, 0, time.UTC),
				},
				criteria: SearchCriteria{
					Authors: &[]string{"Mark Twain"},
				},
			},
			want: false,
		},
		{
			name: "MultipleCriteriaMismatch",
			args: args{
				book: &Book{
					Title:      "To Kill a Mockingbird",
					ISBN:       "5556677788",
					Authors:    []string{"Harper Lee"},
					Publishers: []string{"J.B. Lippincott & Co."},
					Genres:     []string{"Fiction"},
					Pages:      281,
					Edition:    1,
					Published:  time.Date(1960, time.July, 11, 0, 0, 0, 0, time.UTC),
				},
				criteria: SearchCriteria{
					ISBN:       ptrStr("9999999999"),
					Borrowed:   ptrBool(true),
					MinPages:   ptrInt(300),
					MaxPages:   ptrInt(400),
					MinEdition: ptrInt(2),
					MaxEdition: ptrInt(3),
				},
			},
			want: false,
		},
		{
			name: "TitleAndPublisherMatchOnly",
			args: args{
				book: &Book{
					Title:      "Pride and Prejudice",
					ISBN:       "2223334445",
					Authors:    []string{"Jane Austen"},
					Publishers: []string{"T. Egerton"},
					Genres:     []string{"Romance", "Fiction"},
					Pages:      432,
					Edition:    1,
					Published:  time.Date(1813, time.January, 28, 0, 0, 0, 0, time.UTC),
				},
				criteria: SearchCriteria{
					Title:      ptrStr("Pride and Prejudice"),
					Publishers: &[]string{"T. Egerton"},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesAllCriteria(tt.args.book, tt.args.criteria); got != tt.want {
				t.Errorf("matchesAllCriteria() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ptrStr is a helper function for creating a string pointer.
func ptrStr(s string) *string {
	return &s
}

// ptrBool is a helper function for creating a boolean pointer.
func ptrBool(b bool) *bool {
	return &b
}

// ptrInt is a helper function for creating an int pointer.
func ptrInt(i int) *int {
	return &i
}

// ptrTime is a helper function for creating a time pointer.
func ptrTime(t time.Time) *time.Time {
	return &t
}

// ptrStrSlice is a helper function for creating a slice pointer.
func ptrStrSlice(s []string) *[]string {
	return &s
}
