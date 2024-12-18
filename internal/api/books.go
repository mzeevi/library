package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/data"
	"time"
)

type GetBookInput struct {
	ID string `json:"id" path:"id"`
}

type GetBookOutput struct {
	Body data.Book
}

type GetBooksInput struct {
	PaginationInput
	Sort string `json:"sort,omitempty" query:"sort" enum:"id,pages,edition,copies,borrowedCopies,publishedAt,title,isbn,-id,-pages,-edition,-copies,-borrowedCopies,-publishedAt,-title,-isbn"`
}

type GetBooksOutput struct {
	Body BooksInfo
}

type BooksInfo struct {
	Books    []data.Book   `json:"books"`
	Metadata data.Metadata `json:"metadata"`
}

type CreateBookInput struct {
	Body struct {
		Pages       int       `json:"pages" minimum:"1"`
		Edition     int       `json:"edition" minimum:"1"`
		Copies      int       `json:"copies" minimum:"1"`
		PublishedAt time.Time `json:"published_at" format:"date-time"`
		Title       string    `json:"title" minLength:"1"`
		ISBN        string    `json:"isbn" minLength:"13" maxLength:"13"`
		Authors     []string  `json:"authors" minItems:"1" uniqueItems:"true"`
		Publishers  []string  `json:"publishers" minItems:"1" uniqueItems:"true"`
		Genres      []string  `json:"genres" minItems:"1" uniqueItems:"true"`
	}
}

type CreateBookOutput struct {
	Location string    `header:"Location"`
	Body     data.Book `json:"book"`
}

type UpdateBookInput struct {
	ID   string `json:"id" path:"id"`
	Body struct {
		Pages       *int       `json:"pages,omitempty" minimum:"1"`
		Edition     *int       `json:"edition,omitempty" minimum:"1"`
		Copies      *int       `json:"copies,omitempty"  minimum:"1"`
		PublishedAt *time.Time `json:"published_at,omitempty" format:"date-time"`
		Title       *string    `json:"title,omitempty" minLength:"1"`
		ISBN        *string    `json:"isbn,omitempty" minLength:"13" maxLength:"13"`
		Authors     []string   `json:"authors,omitempty" minItems:"1" uniqueItems:"true"`
		Publishers  []string   `json:"publishers,omitempty" minItems:"1" uniqueItems:"true"`
		Genres      []string   `json:"genres,omitempty" minItems:"1" uniqueItems:"true"`
	}
}

type UpdateBookOutput struct {
	Body data.Book `json:"book"`
}

type DeleteBookInput struct {
	ID string `json:"id" path:"id"`
}

type DeleteBookOutput struct {
	Body string `json:"message"`
}

func (b *GetBookInput) Resolve(ctx huma.Context) []error {
	return []error{
		validateID(&b.ID, "path.id"),
	}
}

func (b *UpdateBookInput) Resolve(ctx huma.Context) []error {
	return []error{
		validateID(&b.ID, "path.id"),
	}
}

func (b *DeleteBookInput) Resolve(ctx huma.Context) []error {
	return []error{
		validateID(&b.ID, "path.id"),
	}
}

// getBookHandler retrieves a book by its ID.
func (app *Application) getBookHandler(ctx context.Context, input *GetBookInput) (*GetBookOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	book, err := app.models.Books.Get(ctx, data.BookFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &GetBookOutput{}, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &GetBookOutput{}, err
		}
	}

	resp := &GetBookOutput{}
	resp.Body = *book

	return resp, nil
}

// getBooksHandler retrieves a paginated list of books with sorting options.
func (app *Application) getBooksHandler(ctx context.Context, input *GetBooksInput) (*GetBooksOutput, error) {
	paginator := data.Paginator{Page: input.Page, PageSize: input.PageSize}
	filter := data.BookFilter{}
	sorter := data.Sorter{Field: input.Sort, SortSafelist: supportedBooksSortFields}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	books, metadata, err := app.models.Books.GetAll(ctx, filter, paginator, sorter)
	if err != nil {
		return &GetBooksOutput{}, err
	}

	resp := &GetBooksOutput{
		Body: BooksInfo{
			Books:    books,
			Metadata: metadata,
		},
	}

	return resp, nil
}

// createBookHandler creates a new book record.
func (app *Application) createBookHandler(ctx context.Context, input *CreateBookInput) (*CreateBookOutput, error) {
	book := &data.Book{
		Title:       input.Body.Title,
		ISBN:        input.Body.ISBN,
		Copies:      input.Body.Copies,
		PublishedAt: input.Body.PublishedAt,
		Authors:     input.Body.Authors,
		Genres:      input.Body.Genres,
		Publishers:  input.Body.Publishers,
		Edition:     input.Body.Edition,
		Pages:       input.Body.Pages,
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	id, err := app.models.Books.Insert(ctx, book)
	if err != nil {
		return &CreateBookOutput{}, err
	}

	resp := &CreateBookOutput{
		Body:     *book,
		Location: fmt.Sprintf("%s/%s/%s", basePath, booksKey, id),
	}

	return resp, nil
}

// updateBookHandler updates an existing book record by its ID.
func (app *Application) updateBookHandler(ctx context.Context, input *UpdateBookInput) (*UpdateBookOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	book, err := app.models.Books.Get(ctx, data.BookFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return nil, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &UpdateBookOutput{}, err
		}
	}

	if input.Body.Title != nil {
		book.Title = *input.Body.Title
	}

	if input.Body.ISBN != nil {
		book.ISBN = *input.Body.ISBN
	}

	if input.Body.Pages != nil {
		book.Pages = *input.Body.Pages
	}

	if input.Body.Edition != nil {
		book.Edition = *input.Body.Edition
	}

	if input.Body.Copies != nil {
		book.Copies = *input.Body.Copies
	}

	if input.Body.PublishedAt != nil {
		book.PublishedAt = *input.Body.PublishedAt
	}

	if input.Body.Publishers != nil {
		book.Publishers = input.Body.Publishers
	}

	if input.Body.Genres != nil {
		book.Genres = input.Body.Genres
	}

	if input.Body.Authors != nil {
		book.Authors = input.Body.Authors
	}

	err = app.models.Books.Update(ctx, data.BookFilter{ID: &input.ID}, book)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			return &UpdateBookOutput{}, huma.Error409Conflict(errConflictMsg)
		default:
			return &UpdateBookOutput{}, err
		}
	}

	resp := &UpdateBookOutput{
		Body: *book,
	}

	return resp, nil
}

// deleteBookHandler deletes a book by its ID.
func (app *Application) deleteBookHandler(ctx context.Context, input *DeleteBookInput) (*DeleteBookOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	err := app.models.Books.Delete(ctx, data.BookFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &DeleteBookOutput{}, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &DeleteBookOutput{}, err
		}
	}

	resp := &DeleteBookOutput{
		Body: "book successfully deleted",
	}

	return resp, nil
}
