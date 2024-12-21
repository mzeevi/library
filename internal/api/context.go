package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/data"
)

type contextKey string

const (
	adminContextKey  = contextKey("admin")
	patronContextKey = contextKey("patron")
)

// contextSetPatron adds the Patron to the context.
func (app *Application) contextSetPatron(ctx huma.Context, patron *data.Patron) huma.Context {
	ctx = huma.WithValue(ctx, patronContextKey, patron)
	return ctx
}

// contextGetPatron gets the Admin from the context.
func (app *Application) contextGetPatron(ctx huma.Context) (*data.Patron, bool) {
	patron, ok := ctx.Context().Value(patronContextKey).(*data.Patron)
	return patron, ok
}

// contextSetAdmin adds the Admin to the context.
func (app *Application) contextSetAdmin(ctx huma.Context, admin *data.Admin) huma.Context {
	ctx = huma.WithValue(ctx, adminContextKey, admin)
	return ctx
}

// contextGetAdmin gets the Admin from the context.
func (app *Application) contextGetAdmin(ctx huma.Context) (*data.Admin, bool) {
	admin, ok := ctx.Context().Value(adminContextKey).(*data.Admin)
	return admin, ok
}
