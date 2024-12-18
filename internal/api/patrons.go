package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/data"
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
		Category string `json:"category" enum:"teacher,student"`
	}
}

type CreatePatronOutput struct {
	Location string      `header:"Location"`
	Body     data.Patron `json:"book"`
}

type UpdatePatronInput struct {
	ID   string `json:"id" path:"id"`
	Body struct {
		Name     *string `json:"name,omitempty" minLength:"1"`
		Email    *string `json:"email,omitempty"`
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

// Resolve validates the input in CreatePatronInput.
func (p *CreatePatronInput) Resolve(ctx huma.Context) []error {
	return []error{validateEmail(&p.Body.Email)}
}

// Resolve validates the input in UpdatePatronInput.
func (p *UpdatePatronInput) Resolve(ctx huma.Context) []error {
	return []error{validateEmail(p.Body.Email)}
}

// getPatronHandler retrieves a single patron by ID.
func (app *Application) getPatronHandler(ctx context.Context, input *GetPatronInput) (*GetPatronOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	patron, err := app.models.Patrons.Get(ctx, data.PatronFilter{ID: &input.ID})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &GetPatronOutput{}, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &GetPatronOutput{}, err
		}
	}

	patronTransactions, _, err := app.models.Transactions.GetAll(ctx, data.TransactionFilter{PatronID: &input.ID}, data.Paginator{}, data.Sorter{})
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

	patrons, metadata, err := app.models.Patrons.GetAll(ctx, filter, paginator, sorter)
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

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	id, err := app.models.Patrons.Insert(ctx, patron)
	if err != nil {
		return &CreatePatronOutput{}, err
	}

	resp := &CreatePatronOutput{
		Body:     *patron,
		Location: fmt.Sprintf("%s/%s/%s", basePath, patronsKey, id),
	}

	return resp, nil
}

// updatePatronHandler updates an existing patron based on the provided ID and fields.
func (app *Application) updatePatronHandler(ctx context.Context, input *UpdatePatronInput) (*UpdatePatronOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	patron, err := app.models.Patrons.Get(ctx, data.PatronFilter{ID: &input.ID})
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

	err = app.models.Patrons.Update(ctx, data.PatronFilter{ID: &input.ID}, patron)
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

	err := app.models.Patrons.Delete(ctx, data.PatronFilter{ID: &input.ID})
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
