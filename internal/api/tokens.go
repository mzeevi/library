package api

import (
	"context"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/auth"
	"github.com/mzeevi/library/internal/data"
)

const (
	errInvalidAuthenticationCreds = "invalid authentication credentials"
)

type CreateAuthTokenInput struct {
	Body struct {
		Email    string `json:"email"`
		Password string `json:"password" minLength:"8" maxLength:"72"`
	}
}

type CreateAuthTokenOutput struct {
	Body TokenInfo
}

type TokenInfo struct {
	AuthToken string `json:"auth_token"`
}

// Resolve validates the input in CreatePatronInput.
func (p *CreateAuthTokenInput) Resolve(ctx huma.Context) []error {
	var errs []error

	err := validateEmail(&p.Body.Email, "body.email")
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// createAuthTokenHandler creates and authentication token.
func (app *Application) createAuthTokenHandler(ctx context.Context, input *CreateAuthTokenInput) (*CreateAuthTokenOutput, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	patron, err := app.Models.Patrons.Get(ctx, data.PatronFilter{Email: &input.Body.Email})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDocumentNotFound):
			return &CreateAuthTokenOutput{}, huma.Error404NotFound(errNotFoundMsg)
		default:
			return &CreateAuthTokenOutput{}, err
		}
	}

	match, err := patron.Password.Matches(input.Body.Password)
	if err != nil {
		return &CreateAuthTokenOutput{}, err
	}

	if !match {
		return &CreateAuthTokenOutput{}, huma.Error401Unauthorized(errInvalidAuthenticationCreds)
	}

	jwtBytes, err := auth.CreateJWT(patron.ID, app.Config.JTW.Secret, app.Config.JTW.Issuer, app.Config.JTW.Audience)
	if err != nil {
		return &CreateAuthTokenOutput{}, err
	}

	resp := &CreateAuthTokenOutput{
		Body: TokenInfo{
			AuthToken: string(jwtBytes),
		},
	}

	return resp, nil

}
