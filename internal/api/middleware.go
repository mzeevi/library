package api

import (
	"encoding/base64"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/mzeevi/library/internal/data"
	"github.com/pascaldekloe/jwt"
	"net/http"
	"slices"
	"strings"
	"time"
)

const (
	errInvalidTokenMsg           = "invalid or missing authentication token"
	errInternalServerErrorMsg    = "the server encountered a problem and could not process your request"
	errAuthenticationRequiredMsg = "you must be authenticated to access this resource"
	errInActiveAccountMsg        = "your user account must be activated to access this resource"
	errNotPermittedMsg           = "your account doesn't have the necessary permissions to access this resource"
)

const (
	headerWWWAuthenticateKey = "WWW-Authenticate"
	headerAuthorizationKey   = "Authorization"
	bearerKey                = "Bearer"
)

// authenticate handles both JWT and Basic authentication by dynamically detecting the authType.
func (app *Application) authenticate(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		ctx.SetHeader("Vary", headerAuthorizationKey)

		authHeader := ctx.Header(headerAuthorizationKey)
		if authHeader == "" {
			ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
			return
		}

		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 {
			ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
			return
		}

		switch authParts[0] {
		case "Bearer":
			token := authParts[1]

			claims, err := jwt.HMACCheck([]byte(token), []byte(app.Config.JTW.Secret))
			if err != nil || !claims.Valid(time.Now()) || claims.Issuer != app.Config.JTW.Issuer || !claims.AcceptAudience(app.Config.JTW.Audience) {
				ctx.SetHeader(headerWWWAuthenticateKey, bearerKey)
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
				return
			}

			if !claims.Valid(time.Now()) {
				ctx.SetHeader(headerWWWAuthenticateKey, bearerKey)
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
				return
			}

			if claims.Issuer != app.Config.JTW.Issuer {
				ctx.SetHeader(headerWWWAuthenticateKey, bearerKey)
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
				return
			}

			if !claims.AcceptAudience(app.Config.JTW.Audience) {
				ctx.SetHeader(headerWWWAuthenticateKey, bearerKey)
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
				return
			}

			patron, err := app.Models.Patrons.Get(ctx.Context(), data.PatronFilter{ID: &claims.Subject})
			if err != nil {
				switch {
				case errors.Is(err, data.ErrDocumentNotFound):
					ctx.SetHeader(headerWWWAuthenticateKey, bearerKey)
					_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
				default:
					_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, errInternalServerErrorMsg)
				}
			}

			ctx = app.contextSetPatron(ctx, patron)
		case "Basic":
			credentials, err := base64.StdEncoding.DecodeString(authParts[1])
			if err != nil {
				ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
				return
			}

			creds := strings.SplitN(string(credentials), ":", 2)
			if len(creds) != 2 {
				ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
				return
			}

			username := creds[0]
			password := creds[1]

			admin, err := app.Models.Admins.Get(ctx.Context(), data.AdminFilter{Name: &username})
			if err != nil {
				switch {
				case errors.Is(err, data.ErrDocumentNotFound):
					ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
					_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
					return
				default:
					ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
					_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, errInternalServerErrorMsg)
					return
				}
			}

			matches, err := admin.Password.Matches(password)
			if err != nil {
				ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
				_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, errInternalServerErrorMsg)
			}

			if !matches {
				ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
				return
			}

			ctx = app.contextSetAdmin(ctx, admin)
		default:
			ctx.SetHeader(headerWWWAuthenticateKey, `Basic realm="Restricted"`)
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errInvalidTokenMsg)
			return
		}

		next(ctx)
	}
}

// requireAuthenticatedPatron ensures the request is made by an authenticated patron.
func (app *Application) requireAuthenticatedPatron(api huma.API, inFn func(ctx huma.Context, next func(huma.Context))) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if admin, ok := app.contextGetAdmin(ctx); ok {
			if !admin.IsAnonymous() {
				inFn(ctx, next)
				return
			}
		}

		if patron, ok := app.contextGetPatron(ctx); ok {
			if !patron.IsAnonymous() {
				inFn(ctx, next)
				return
			}
		}

		_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, errAuthenticationRequiredMsg)
	}
}

// requireActivatedPatron ensures the request is made by an authenticated and activated patron.
func (app *Application) requireActivatedPatron(api huma.API, inFn func(ctx huma.Context, next func(huma.Context))) func(ctx huma.Context, next func(huma.Context)) {
	fn := func(ctx huma.Context, next func(huma.Context)) {
		if admin, ok := app.contextGetAdmin(ctx); ok {
			if admin.Activated {
				inFn(ctx, next)
				return
			}
		}

		if patron, ok := app.contextGetPatron(ctx); ok {
			if patron.Activated {
				inFn(ctx, next)
				return
			}
		}

		_ = huma.WriteErr(api, ctx, http.StatusForbidden, errInActiveAccountMsg)
	}

	return app.requireAuthenticatedPatron(api, fn)
}

// requirePermission dynamically checks if the authenticated user (patron or admin) has the required permission.
func (app *Application) requirePermission(api huma.API, code string) func(ctx huma.Context, next func(huma.Context)) {
	fn := func(ctx huma.Context, next func(huma.Context)) {
		if admin, ok := app.contextGetAdmin(ctx); ok {
			if slices.Contains(admin.Permissions, code) {
				next(ctx)
				return
			}
		} else if patron, ok := app.contextGetPatron(ctx); ok {
			if slices.Contains(patron.Permissions, code) {
				next(ctx)
				return
			}
		} else {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, errNotPermittedMsg)
		}
	}

	return app.requireActivatedPatron(api, fn)
}

// requireMatchingID ensures a Patron makes requests only with its own ID.
func (app *Application) requireMatchingID(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		id := ctx.Param(idKey)
		if id != "" {
			if _, ok := app.contextGetAdmin(ctx); ok {
				next(ctx)
				return
			} else if patron, ok := app.contextGetPatron(ctx); ok {
				if patron.ID == id {
					next(ctx)
					return
				} else {
					_ = huma.WriteErr(api, ctx, http.StatusForbidden, errNotPermittedMsg)
					return
				}
			} else {
				_ = huma.WriteErr(api, ctx, http.StatusForbidden, errNotPermittedMsg)
				return
			}
		}

		_ = huma.WriteErr(api, ctx, http.StatusForbidden, errNotPermittedMsg)
	}
}
