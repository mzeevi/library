package data

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type TestCategory struct {
	DiscountPercentage float64
}

func (t TestCategory) Discount() float64 {
	return t.DiscountPercentage / 100
}

func TestNewPatron(t *testing.T) {
	discounts := map[PatronCategoryType]float64{
		Teacher: 15.0,
		Student: 10.0,
	}

	tests := []struct {
		name string
		args struct {
			patronName    string
			categoryType  PatronCategoryType
			discountRates map[PatronCategoryType]float64
		}
		wantErr      error
		wantDiscount float64
	}{
		{
			name: "ValidTeacherPatron",
			args: struct {
				patronName    string
				categoryType  PatronCategoryType
				discountRates map[PatronCategoryType]float64
			}{
				patronName:    "John Doe",
				categoryType:  Teacher,
				discountRates: discounts,
			},
			wantErr:      nil,
			wantDiscount: 15.0,
		},
		{
			name: "ValidStudentPatron",
			args: struct {
				patronName    string
				categoryType  PatronCategoryType
				discountRates map[PatronCategoryType]float64
			}{
				patronName:    "Jane Smith",
				categoryType:  Student,
				discountRates: discounts,
			},
			wantErr:      nil,
			wantDiscount: 10.0,
		},
		{
			name: "UnknownPatronCategory",
			args: struct {
				patronName    string
				categoryType  PatronCategoryType
				discountRates map[PatronCategoryType]float64
			}{
				patronName:    "Unknown Patron",
				categoryType:  PatronCategoryType("999"), // Invalid category
				discountRates: discounts,
			},
			wantErr:      errors.New(errUnknownCategory),
			wantDiscount: 0.0,
		},
		{
			name: "MissingDiscountForTeacher",
			args: struct {
				patronName    string
				categoryType  PatronCategoryType
				discountRates map[PatronCategoryType]float64
			}{
				patronName:    "John Doe",
				categoryType:  Teacher,
				discountRates: map[PatronCategoryType]float64{Student: 10.0}, // Missing Teacher discount
			},
			wantErr:      nil,
			wantDiscount: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPatron, err := NewPatron(tt.args.patronName, tt.args.categoryType, tt.args.discountRates)

			if (err != nil && tt.wantErr == nil) || (err == nil && tt.wantErr != nil) || (err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error()) {
				t.Errorf("NewPatron() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if err == nil && gotPatron.Name != tt.args.patronName {
				t.Errorf("NewPatron() Name = %v, want %v", gotPatron.Name, tt.args.patronName)
			}

			if err == nil {
				switch category := gotPatron.Category.(type) {
				case TeacherCategory:
					if category.DiscountPercentage != tt.wantDiscount {
						t.Errorf("NewPatron() Discount = %v, want %v", category.DiscountPercentage, tt.wantDiscount)
					}
				case StudentCategory:
					if category.DiscountPercentage != tt.wantDiscount {
						t.Errorf("NewPatron() Discount = %v, want %v", category.DiscountPercentage, tt.wantDiscount)
					}
				default:
					if tt.wantDiscount != 0.0 {
						t.Errorf("NewPatron() Discount = %v, want %v", 0.0, tt.wantDiscount)
					}
				}
			}
		})
	}
}

func TestBorrowBook(t *testing.T) {
	tests := []struct {
		name             string
		bookList         []*Book
		patron           Patron
		titleToBorrow    string
		expectedError    error
		expectedBorrowed bool
	}{
		{
			name: "SuccessfulBorrow",
			bookList: []*Book{
				{Title: "Go Programming", ISBN: "123456", Borrowed: false, BorrowDuration: 7},
			},
			patron:           Patron{Name: "John Doe"},
			titleToBorrow:    "Go Programming",
			expectedError:    nil,
			expectedBorrowed: true,
		},
		{
			name: "BookAlreadyBorrowed",
			bookList: []*Book{
				{Title: "Go Programming", ISBN: "123456", Borrowed: true, BorrowDuration: 7},
			},
			patron:           Patron{Name: "John Doe"},
			titleToBorrow:    "Go Programming",
			expectedError:    fmt.Errorf("%v: %v", errUnableToBorrow, errBookAlreadyBorrowed),
			expectedBorrowed: true,
		},
		{
			name: "BookNotFound",
			bookList: []*Book{
				{Title: "Go Programming", ISBN: "123456", Borrowed: false, BorrowDuration: 7},
			},
			patron:           Patron{Name: "John Doe"},
			titleToBorrow:    "Nonexistent Book",
			expectedError:    fmt.Errorf("%v: %v", errUnableToBorrow, errNonexistentBook),
			expectedBorrowed: false,
		},
		{
			name: "NilBorrowedBooksMap",
			bookList: []*Book{
				{Title: "Go Programming", ISBN: "123456", Borrowed: false, BorrowDuration: 7},
			},
			patron:           Patron{Name: "John Doe", BorrowedBooks: nil},
			titleToBorrow:    "Go Programming",
			expectedError:    nil,
			expectedBorrowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.patron.BorrowBook(tt.titleToBorrow, tt.bookList)

			if err != nil && tt.expectedError == nil {
				t.Errorf("Expected no error, but got %v", err)
			} else if err == nil && tt.expectedError != nil {
				t.Errorf("Expected error %v, but got none", tt.expectedError)
			} else if err != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("Expected error %v, but got %v", tt.expectedError, err)
			}

			if err == nil && !tt.patron.BorrowedBooks[tt.titleToBorrow].BorrowedAt.IsZero() != tt.expectedBorrowed {
				t.Errorf("Expected book to be borrowed: %v, but got %v", tt.expectedBorrowed, tt.patron.BorrowedBooks[tt.titleToBorrow].BorrowedAt.IsZero())
			}
		})
	}
}

func TestCalcFine(t *testing.T) {
	type args struct {
		overdueFine float64
	}
	tests := []struct {
		name   string
		patron *Patron
		args   args
		want   float64
	}{
		{
			name: "CalculateFineWithSingleBook",
			patron: &Patron{
				Name: "John Doe",
				Category: TestCategory{
					DiscountPercentage: 10,
				},
				BorrowedBooks: map[string]bookDetails{
					"Book 1": {
						ISBN:           "1234567890",
						BorrowDuration: 7 * 24 * time.Hour,
						BorrowedAt:     time.Now().Add(-10 * 24 * time.Hour),
					},
				},
			},
			args: args{
				overdueFine: 15,
			},
			want: 54,
		},
		{
			name: "CalculateFineWithMultipleBooks",
			patron: &Patron{
				Name: "Jane Smith",
				Category: TestCategory{
					DiscountPercentage: 20,
				},
				BorrowedBooks: map[string]bookDetails{
					"Book 1": {
						ISBN:           "1234567890",
						BorrowDuration: 5 * 24 * time.Hour,
						BorrowedAt:     time.Now().Add(-8 * 24 * time.Hour),
					},
					"Book 2": {
						ISBN:           "0987654321",
						BorrowDuration: 3 * 24 * time.Hour,
						BorrowedAt:     time.Now().Add(-7 * 24 * time.Hour),
					},
				},
			},
			args: args{
				overdueFine: 2.0,
			},
			want: 14.4,
		},
		{
			name: "CalculateFineNoBooks",
			patron: &Patron{
				Name: "Mark Lee",
				Category: TestCategory{
					DiscountPercentage: 15,
				},
				BorrowedBooks: map[string]bookDetails{},
			},
			args: args{
				overdueFine: 2.0,
			},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.patron.CalcFine(tt.args.overdueFine)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalcFine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBorrowedBooks(t *testing.T) {
	zeroTime := time.Now()

	tests := []struct {
		name   string
		patron *Patron
		want   map[string]time.Time
	}{
		{
			name: "SingleBookBorrowed",
			patron: &Patron{
				Name: "John Doe",
				BorrowedBooks: map[string]bookDetails{
					"Book 1": {
						ISBN:           "1234567890",
						BorrowDuration: 7 * 24 * time.Hour,
						BorrowedAt:     zeroTime.Add(-10 * 24 * time.Hour),
					},
				},
			},
			want: map[string]time.Time{
				"Book 1": zeroTime.Add(-3 * 24 * time.Hour),
			},
		},
		{
			name: "MultipleBooksBorrowed",
			patron: &Patron{
				Name: "Jane Smith",
				BorrowedBooks: map[string]bookDetails{
					"Book 1": {
						ISBN:           "1234567890",
						BorrowDuration: 5 * 24 * time.Hour,
						BorrowedAt:     zeroTime.Add(-8 * 24 * time.Hour),
					},
					"Book 2": {
						ISBN:           "0987654321",
						BorrowDuration: 3 * 24 * time.Hour,
						BorrowedAt:     zeroTime.Add(-7 * 24 * time.Hour),
					},
				},
			},
			want: map[string]time.Time{
				"Book 1": zeroTime.Add(-3 * 24 * time.Hour),
				"Book 2": zeroTime.Add(-4 * 24 * time.Hour),
			},
		},
		{
			name: "NoBooksBorrowed",
			patron: &Patron{
				Name:          "Mark Lee",
				BorrowedBooks: map[string]bookDetails{},
			},
			want: map[string]time.Time{},
		},
	}

	// Iterate over the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.patron.GetBorrowedBooks()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBorrowedBooks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReturnBook(t *testing.T) {
	type args struct {
		title string
		books []*Book
	}
	tests := []struct {
		name    string
		args    args
		patron  *Patron
		wantErr error
	}{
		{
			name: "ReturnBookSuccessfully",
			args: args{
				title: "Test Book",
				books: []*Book{
					{Title: "Test Book", ISBN: "1234567890", Borrowed: true},
				},
			},
			patron: &Patron{
				Name: "John Doe",
				BorrowedBooks: map[string]bookDetails{
					"Test Book": {ISBN: "1234567890", BorrowDuration: time.Hour},
				},
			},
			wantErr: nil,
		},
		{
			name: "BookNotFound",
			args: args{
				title: "Nonexistent Book",
				books: []*Book{
					{Title: "Test Book", ISBN: "1234567890", Borrowed: true},
				},
			},
			patron: &Patron{
				Name: "John Doe",
				BorrowedBooks: map[string]bookDetails{
					"Test Book": {ISBN: "1234567890", BorrowDuration: time.Hour},
				},
			},
			wantErr: fmt.Errorf("%v: %v", errUnableToReturn, errors.New(errNonexistentBook)),
		},
		{
			name: "BookNotOwnedByPatron",
			args: args{
				title: "Another Book",
				books: []*Book{
					{Title: "Test Book", ISBN: "1234567890", Borrowed: true},
					{Title: "Another Book", ISBN: "0987654321", Borrowed: true},
				},
			},
			patron: &Patron{
				Name: "John Doe",
				BorrowedBooks: map[string]bookDetails{
					"Test Book": {ISBN: "1234567890", BorrowDuration: time.Hour},
				},
			},
			wantErr: fmt.Errorf("%v: %v", errUnableToReturn, errors.New(errBookNotOwned)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.patron.ReturnBook(tt.args.title, tt.args.books)

			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("ReturnBook() error = %v, wantErr %v", gotErr, tt.wantErr)
			}

			if tt.wantErr == nil {
				if _, exists := tt.patron.BorrowedBooks[tt.args.title]; exists {
					t.Errorf("ReturnBook() failed to remove book from borrowed books")
				}
			}
		})
	}
}

func TestUpdatePatron(t *testing.T) {
	discounts := map[PatronCategoryType]float64{
		Teacher: 15.0,
		Student: 10.0,
	}

	tests := []struct {
		name string
		args struct {
			name        *string
			category    *PatronCategoryType
			discountMap map[PatronCategoryType]float64
		}
		initial struct {
			name      string
			category  PatronCategoryType
			discounts map[PatronCategoryType]float64
		}
		wantErr      error
		wantName     string
		wantDiscount float64
	}{
		{
			name: "NameOnly",
			args: struct {
				name        *string
				category    *PatronCategoryType
				discountMap map[PatronCategoryType]float64
			}{
				name:        ptrStr("Jane Smith"),
				category:    nil,
				discountMap: discounts,
			},
			initial: struct {
				name      string
				category  PatronCategoryType
				discounts map[PatronCategoryType]float64
			}{
				name:      "John Doe",
				category:  Teacher,
				discounts: discounts,
			},
			wantErr:      nil,
			wantName:     "Jane Smith",
			wantDiscount: 15.0,
		},
		{
			name: "CategoryOnly",
			args: struct {
				name        *string
				category    *PatronCategoryType
				discountMap map[PatronCategoryType]float64
			}{
				name:        nil,
				category:    ptrPatronCategory(Student),
				discountMap: discounts,
			},
			initial: struct {
				name      string
				category  PatronCategoryType
				discounts map[PatronCategoryType]float64
			}{
				name:      "John Doe",
				category:  Teacher,
				discounts: discounts,
			},
			wantErr:      nil,
			wantName:     "John Doe",
			wantDiscount: 10.0,
		},
		{
			name: "NameAndCategory",
			args: struct {
				name        *string
				category    *PatronCategoryType
				discountMap map[PatronCategoryType]float64
			}{
				name:        ptrStr("Jane Smith"),
				category:    ptrPatronCategory(Student),
				discountMap: discounts,
			},
			initial: struct {
				name      string
				category  PatronCategoryType
				discounts map[PatronCategoryType]float64
			}{
				name:      "John Doe",
				category:  Teacher,
				discounts: discounts,
			},
			wantErr:      nil,
			wantName:     "Jane Smith",
			wantDiscount: 10.0,
		},
		{
			name: "UnknownCategory",
			args: struct {
				name        *string
				category    *PatronCategoryType
				discountMap map[PatronCategoryType]float64
			}{
				name:        nil,
				category:    ptrPatronCategory("999"),
				discountMap: discounts,
			},
			initial: struct {
				name      string
				category  PatronCategoryType
				discounts map[PatronCategoryType]float64
			}{
				name:      "John Doe",
				category:  Teacher,
				discounts: discounts,
			},
			wantErr:      errors.New(errUnknownCategory),
			wantName:     "John Doe",
			wantDiscount: 15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patron := Patron{
				ID:        1,
				Name:      tt.initial.name,
				Category:  TeacherCategory{DiscountPercentage: tt.initial.discounts[Teacher]},
				CreatedAt: time.Now(),
			}

			err := patron.UpdatePatron(tt.args.name, tt.args.category, tt.args.discountMap)

			if (err != nil && tt.wantErr == nil) || (err == nil && tt.wantErr != nil) || (err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error()) {
				t.Errorf("UpdatePatron() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if patron.Name != tt.wantName {
				t.Errorf("UpdatePatron() Name = %v, want %v", patron.Name, tt.wantName)
			}

			if err == nil {
				switch category := patron.Category.(type) {
				case TeacherCategory:
					if category.DiscountPercentage != tt.wantDiscount {
						t.Errorf("UpdatePatron() Discount = %v, want %v", category.DiscountPercentage, tt.wantDiscount)
					}
				case StudentCategory:
					if category.DiscountPercentage != tt.wantDiscount {
						t.Errorf("UpdatePatron() Discount = %v, want %v", category.DiscountPercentage, tt.wantDiscount)
					}
				default:
					if tt.wantDiscount != 0.0 {
						t.Errorf("UpdatePatron() Discount = %v, want %v", 0.0, tt.wantDiscount)
					}
				}
			}
		})
	}
}

func TestStudentCategoryDiscount(t *testing.T) {
	tests := []struct {
		name     string
		category StudentCategory
		expected float64
	}{
		{
			name:     "StudentCategoryDiscount10%",
			category: StudentCategory{DiscountPercentage: 10.0},
			expected: 0.10,
		},
		{
			name:     "StudentCategoryDiscount0%",
			category: StudentCategory{DiscountPercentage: 0.0},
			expected: 0.0,
		},
		{
			name:     "StudentCategoryDiscountNegative",
			category: StudentCategory{DiscountPercentage: -3.0},
			expected: -0.03,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.category.Discount()
			if got != tt.expected {
				t.Errorf("Discount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTeacherCategoryDiscount(t *testing.T) {
	tests := []struct {
		name     string
		category TeacherCategory
		expected float64
	}{
		{
			name:     "TeacherCategoryDiscount20%",
			category: TeacherCategory{DiscountPercentage: 20.0},
			expected: 0.20,
		},
		{
			name:     "TeacherCategoryDiscount0%",
			category: TeacherCategory{DiscountPercentage: 0.0},
			expected: 0.0,
		},
		{
			name:     "TeacherCategoryDiscountNegative",
			category: TeacherCategory{DiscountPercentage: -5.0},
			expected: -0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.category.Discount()
			if got != tt.expected {
				t.Errorf("Discount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDaysBetween(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			start time.Time
			end   time.Time
		}
		expected int
	}{
		{
			name: "SameDay",
			args: struct {
				start, end time.Time
			}{
				start: time.Date(2024, time.December, 7, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.December, 7, 23, 59, 59, 0, time.UTC),
			},
			expected: 1,
		},
		{
			name: "OneDayApart",
			args: struct {
				start, end time.Time
			}{
				start: time.Date(2024, time.December, 6, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.December, 7, 0, 0, 0, 0, time.UTC),
			},
			expected: 1,
		},
		{
			name: "MultipleDaysApart",
			args: struct {
				start, end time.Time
			}{
				start: time.Date(2024, time.December, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.December, 7, 0, 0, 0, 0, time.UTC),
			},
			expected: 6,
		},
		{
			name: "NegativeDuration",
			args: struct {
				start, end time.Time
			}{
				start: time.Date(2024, time.December, 8, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.December, 7, 0, 0, 0, 0, time.UTC),
			},
			expected: 0,
		},
		{
			name: "VerySmallDuration",
			args: struct {
				start, end time.Time
			}{
				start: time.Date(2024, time.December, 7, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.December, 7, 0, 0, 1, 0, time.UTC),
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := daysBetween(tt.args.start, tt.args.end)
			if got != tt.expected {
				t.Errorf("daysBetween() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func ptrPatronCategory(c PatronCategoryType) *PatronCategoryType {
	return &c
}
