package api

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/auth"
	"github.com/mzeevi/library/internal/data"
	"time"
)

const (
	errInvalidOrExpiredTokenMsg = "invalid or expired activation token"
	errEmailAlreadyExistsMsg    = "a resource with this email address already exists"
)

type GetPatronInput struct {
	ID string `json:"id" path:"id"`
}

type GetPatronOutput struct {
	Body PatronSummary
}

type PatronSummary struct {
	Info         data.Patron         `json:"info"`
	Transactions []patronTransaction `json:"transactions"`
	TotalFine    float64             `json:"total_fine"`
}

type patronTransaction struct {
	Transaction data.Transaction `json:"transaction"`
	Fine        float64          `json:"fine"`
}

type GetPatronsInput struct {
	PaginationInput
	Sort string `json:"sort,omitempty" query:"sort" enum:"category,name,email,-category,-name,-email"`
}

type GetPatronsOutput struct {
	Body PatronsInfo
}

type PatronsInfo struct {
	Patrons  []data.Patron `json:"patrons"`
	Metadata data.Metadata `json:"metadata"`
}

type CreatePatronInput struct {
	Body struct {
		Name     string `json:"name" minLength:"1"`
		Email    string `json:"email"`
		Password string `json:"password" minLength:"8" maxLength:"72"`
		Category string `json:"category" enum:"teacher,student"`
	}
}

type CreatePatronOutput struct {
	Location string        `header:"Location"`
	Body     newPatronInfo `json:"patronInfo"`
}

type newPatronInfo struct {
	Patron data.Patron `json:"patron"`
	Token  string      `json:"token"`
}

type UpdatePatronInput struct {
	ID   string `json:"id" path:"id"`
	Body struct {
		Name     *string `json:"name,omitempty" minLength:"1"`
		Email    *string `json:"email,omitempty"`
		Password *string `json:"password,omitempty" minLength:"8" maxLength:"72"`
		Category *string `json:"category,omitempty" enum:"teacher,student"`
	}
}

type UpdatePatronOutput struct {
	Body data.Patron `json:"patron"`
}

type DeletePatronInput struct {
	ID string `json:"id" path:"id"`
}

type DeletePatronOutput struct {
	Body string `json:"message"`
}

type ActivatePatronInput struct {
	Body struct {
		TokenPlaintext string `json:"token" minLength:"26" maxLength:"26"`
	}
}

type ActivatePatronOutput struct {
	Body data.Patron `json:"patron"`
}

func (p *GetPatronInput) Resolve(ctx huma.Context) []error {
	var errs []error

	err := validateID(&p.ID, "path.id")
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

func (p *DeletePatronInput) Resolve(ctx huma.Context) []error {
	var errs []error

	err := validateID(&p.ID, "path.id")
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Resolve validates the input in CreatePatronInput.
func (p *CreatePatronInput) Resolve(ctx huma.Context) []error {
	var errs []error

	err := validateEmail(&p.Body.Email, "body.email")
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Resolve validates the input in UpdatePatronInput.
func (p *UpdatePatronInput) Resolve(ctx huma.Context) []error {
	var errs []error

	err := validateID(&p.ID, "body.path")
	if err != nil {
		errs = append(errs, err)
	}

	err = validateEmail(p.Body.Email, "body.email")
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// getPatronHandler retrieves a single patron by ID.
func (app *Application) getPatronHandler(ctx context.Context, input *GetPatronInput) (*GetPatronOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	patron, err := app.Models.Patrons.Get(ctx, data.PatronFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &GetPatronOutput{}, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &GetPatronOutput{}, err
		}
	}

	patronTransactions, _, err := app.Models.Transactions.GetAll(ctx, data.TransactionFilter{PatronID: &input.ID}, data.Paginator{}, data.Sorter{})
	if err != nil {
		return &GetPatronOutput{}, err
	}

	transactionsSummary, totalFine := processPatronTransactions(patronTransactions, app.cost.overdueFine)

	resp := &GetPatronOutput{
		Body: PatronSummary{
			Info:         *patron,
			Transactions: transactionsSummary,
			TotalFine:    totalFine,
		},
	}

	return resp, nil
}

// getPatronsHandler retrieves a list of patrons based on filters, pagination, and sorting.
func (app *Application) getPatronsHandler(ctx context.Context, input *GetPatronsInput) (*GetPatronsOutput, error) {
	paginator := data.Paginator{Page: input.Page, PageSize: input.PageSize}
	filter := data.PatronFilter{}

	sorter := data.Sorter{Field: input.Sort, SortSafelist: supportedPatronsSortFields}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	patrons, metadata, err := app.Models.Patrons.GetAll(ctx, filter, paginator, sorter)
	if err != nil {
		return &GetPatronsOutput{}, err
	}

	resp := &GetPatronsOutput{
		Body: PatronsInfo{
			Patrons:  patrons,
			Metadata: metadata,
		},
	}

	return resp, nil
}

// createPatronHandler creates a new patron and stores it in the database.
func (app *Application) createPatronHandler(ctx context.Context, input *CreatePatronInput) (*CreatePatronOutput, error) {
	patron := &data.Patron{
		Name:     input.Body.Name,
		Email:    input.Body.Email,
		Category: input.Body.Category,
	}

	if err := patron.Password.Set(input.Body.Password); err != nil {
		return &CreatePatronOutput{}, err
	}

	permissions := []string{auth.WritePatronPermission, auth.ReadPatronPermission, auth.ReadBooksPermission, auth.BorrowBookPermission, auth.ReturnBookPermission}
	patron.Permissions = permissions

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	id, err := app.Models.Patrons.Insert(ctx, patron)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			return &CreatePatronOutput{}, huma.Error422UnprocessableEntity(errEmailAlreadyExistsMsg)
		case errors.Is(err, data.ErrDuplicateID):
			return &CreatePatronOutput{}, huma.Error422UnprocessableEntity(errIDAlreadyExistsMsg)
		}
		return &CreatePatronOutput{}, err
	}

	token, err := app.Models.Tokens.New(ctx, id, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		return &CreatePatronOutput{}, err
	}

	resp := &CreatePatronOutput{
		Body: newPatronInfo{
			Patron: *patron,
			Token:  token.Plaintext,
		},
		Location: fmt.Sprintf("%s/%s/%s", basePath, patronsKey, id),
	}

	return resp, nil
}

// updatePatronHandler updates an existing patron based on the provided ID and fields.
func (app *Application) updatePatronHandler(ctx context.Context, input *UpdatePatronInput) (*UpdatePatronOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	patron, err := app.Models.Patrons.Get(ctx, data.PatronFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return nil, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &UpdatePatronOutput{}, err
		}
	}

	if input.Body.Name != nil {
		patron.Name = *input.Body.Name
	}

	if input.Body.Email != nil {
		patron.Email = *input.Body.Email
	}

	if input.Body.Category != nil {
		patron.Category = *input.Body.Category
	}

	err = app.Models.Patrons.Update(ctx, data.PatronFilter{ID: &input.ID}, patron)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			return &UpdatePatronOutput{}, huma.Error409Conflict(errConflictMsg)
		default:
			return &UpdatePatronOutput{}, err
		}
	}

	resp := &UpdatePatronOutput{
		Body: *patron,
	}

	return resp, nil
}

// deletePatronHandler deletes a patron based on the provided ID.
func (app *Application) deletePatronHandler(ctx context.Context, input *DeletePatronInput) (*DeletePatronOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	err := app.Models.Patrons.Delete(ctx, data.PatronFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &DeletePatronOutput{}, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &DeletePatronOutput{}, err
		}
	}

	resp := &DeletePatronOutput{
		Body: "book successfully deleted",
	}

	return resp, nil
}

// activatePatronHandler activates a patron.
func (app *Application) activatePatronHandler(ctx context.Context, input *ActivatePatronInput) (*ActivatePatronOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(input.Body.TokenPlaintext))

	patronID, err := app.Models.Tokens.GetPatronID(ctx, data.TokenFilter{
		Plaintext: &input.Body.TokenPlaintext,
		Hash:      tokenHash[:],
		Scope:     ptr(data.ScopeActivation),
		MinExpiry: ptr(time.Now()),
	})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &ActivatePatronOutput{}, huma.Error422UnprocessableEntity(errInvalidOrExpiredTokenMsg)
		default:
			return &ActivatePatronOutput{}, err
		}
	}

	patron, err := app.Models.Patrons.Get(ctx, data.PatronFilter{ID: &patronID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &ActivatePatronOutput{}, huma.Error422UnprocessableEntity(errInvalidOrExpiredTokenMsg)
		default:
			return &ActivatePatronOutput{}, err
		}
	}

	patron.Activated = true
	err = app.Models.Patrons.Update(ctx, data.PatronFilter{ID: &patron.ID}, patron)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			return &ActivatePatronOutput{}, huma.Error409Conflict(errConflictMsg)
		default:
			return &ActivatePatronOutput{}, err
		}
	}

	err = app.Models.Tokens.DeleteAllForPatron(ctx, data.TokenFilter{PatronID: &patronID, Scope: ptr(data.ScopeActivation)})
	if err != nil {
		return &ActivatePatronOutput{}, err
	}

	resp := &ActivatePatronOutput{
		Body: *patron,
	}

	return resp, nil
}
