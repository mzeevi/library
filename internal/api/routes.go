package api

import (
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/mzeevi/library/internal/auth"
	"github.com/mzeevi/library/internal/query"
	"net/http"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
	"github.com/go-chi/httprate"
)

const (
	bearerSecKey      = "bearer"
	basicAuthKey      = "basic"
	basePath          = ""
	booksKey          = "books"
	patronsKey        = "patrons"
	transactionsKey   = "transactions"
	tokensKey         = "token"
	authenticationKey = "authentication"
	borrowKey         = "borrow"
	returnKey         = "return"
	searchKey         = "search"
	healthcheckKey    = "healthcheck"
	idKey             = "id"
	activated         = "activated"
)

var (
	typeInt    = reflect.TypeOf(0)
	typeString = reflect.TypeOf("")
	typeTime   = reflect.TypeOf(time.Now())
)

// routes sets up and returns the HTTP handler for the application.
func (app *Application) routes() http.Handler {
	router := chi.NewMux()
	conf := huma.DefaultConfig("My API", "1.0.0")
	conf.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		bearerSecKey: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "jwt",
		},
		basicAuthKey: {
			Type:         "http",
			Scheme:       "basic",
			BearerFormat: "Basic Auth",
		},
	}

	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(httplog.RequestLogger(app.logger))
	router.Use(middleware.Recoverer)
	router.Use(httprate.Limit(100, 10*time.Second, httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint)))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.Config.CORS.TrustedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	api := humachi.New(router, conf)

	app.registerHealthcheck(api)
	app.registerBooks(api)
	app.registerPatrons(api)
	app.registerTransactions(api)
	app.registerSearch(api)
	app.registerToken(api)

	return router
}

// registerHealthcheck registers healthcheck endpoints.
func (app *Application) registerHealthcheck(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "healthcheck",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s", basePath, healthcheckKey),
		Summary:     "Healthcheck",
		Description: "Basic health check",
		Tags:        []string{healthcheckKey},
	}, app.healthcheckHandler)
}

// registerBooks registers book endpoints.
func (app *Application) registerBooks(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-book",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, booksKey, idKey),
		Summary:     "Get a Book",
		Description: "Get a Book from a specific ID",
		Tags:        []string{booksKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadBooksPermission)},
		Security: []map[string][]string{
			{bearerSecKey: {}},
		},
	}, app.getBookHandler)

	huma.Register(api, huma.Operation{
		OperationID: "get-books",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s", basePath, booksKey),
		Summary:     "Get Books",
		Description: "Get all Books",
		Tags:        []string{booksKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadBooksPermission)},
		Security: []map[string][]string{
			{bearerSecKey: {}},
			{basicAuthKey: {}},
		},
	}, app.getBooksHandler)

	huma.Register(api, huma.Operation{
		OperationID: "create-book",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s", basePath, booksKey),
		Summary:     "Create a Book",
		Description: "Create a specific Book",
		Tags:        []string{booksKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.WriteBooksPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.createBookHandler)

	huma.Register(api, huma.Operation{
		OperationID: "update-book",
		Method:      http.MethodPut,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, booksKey, idKey),
		Summary:     "Update a Book",
		Description: "Update a specific Book",
		Tags:        []string{booksKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.WriteBooksPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.updateBookHandler)

	huma.Register(api, huma.Operation{
		OperationID: "delete-book",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, booksKey, idKey),
		Summary:     "Delete a Book",
		Description: "Delete a specific Book",
		Tags:        []string{booksKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.WriteBooksPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.deleteBookHandler)
}

// registerPatrons registers patron endpoints.
func (app *Application) registerPatrons(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-patron",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, patronsKey, idKey),
		Summary:     "Get a Patron",
		Description: "Get a Patron from a specific ID",
		Tags:        []string{patronsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadPatronPermission), app.requireMatchingID(api)},
		Security: []map[string][]string{
			{bearerSecKey: {}},
			{basicAuthKey: {}},
		},
	}, app.getPatronHandler)

	huma.Register(api, huma.Operation{
		OperationID: "get-patrons",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s", basePath, patronsKey),
		Summary:     "Get Patrons",
		Description: "Get all Patrons",
		Tags:        []string{patronsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadPatronsPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.getPatronsHandler)

	huma.Register(api, huma.Operation{
		OperationID: "create-patron",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s", basePath, patronsKey),
		Summary:     "Create a Patron",
		Description: "Create a specific Patron",
		Tags:        []string{patronsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.WritePatronsPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.createPatronHandler)

	huma.Register(api, huma.Operation{
		OperationID: "update-patron",
		Method:      http.MethodPut,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, patronsKey, idKey),
		Summary:     "Update a Patron",
		Description: "Update a specific Patron",
		Tags:        []string{patronsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.WritePatronPermission)},
		Security: []map[string][]string{
			{bearerSecKey: {}},
			{basicAuthKey: {}},
		},
	}, app.updatePatronHandler)

	huma.Register(api, huma.Operation{
		OperationID: "delete-patron",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, patronsKey, idKey),
		Summary:     "Delete a Patron",
		Description: "Delete a specific Patron",
		Tags:        []string{patronsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.WritePatronPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
			{bearerSecKey: {}},
		},
	}, app.deletePatronHandler)

	huma.Register(api, huma.Operation{
		OperationID: "activate-patron",
		Method:      http.MethodPut,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, patronsKey, activated),
		Summary:     "Activate a Patron",
		Description: "Activate a specific Patron",
		Tags:        []string{patronsKey},
	}, app.activatePatronHandler)
}

// registerTransactions registers transaction endpoints.
func (app *Application) registerTransactions(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-transaction",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, transactionsKey, idKey),
		Summary:     "Get a Transaction",
		Description: "Get a Transaction from a specific ID",
		Tags:        []string{transactionsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadTransactionsPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.getTransactionHandler)

	huma.Register(api, huma.Operation{
		OperationID: "get-transactions",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s", basePath, transactionsKey),
		Summary:     "Get Transactions",
		Description: "Get all Transactions",
		Tags:        []string{transactionsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadTransactionsPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.getTransactionsHandler)

	huma.Register(api, huma.Operation{
		OperationID: "borrow-book-transaction",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, transactionsKey, borrowKey),
		Summary:     "Borrow Book",
		Description: "Borrow a Book",
		Tags:        []string{transactionsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.BorrowBookPermission)},
		Security: []map[string][]string{
			{bearerSecKey: {}},
			{basicAuthKey: {}},
		},
	}, app.borrowBookTransactionHandler)

	huma.Register(api, huma.Operation{
		OperationID: "return-book-transaction",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, transactionsKey, returnKey),
		Summary:     "Return Book",
		Description: "Return a Book",
		Tags:        []string{transactionsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReturnBookPermission)},
		Security: []map[string][]string{
			{bearerSecKey: {}},
			{basicAuthKey: {}},
		},
	}, app.returnBookTransactionHandler)

	huma.Register(api, huma.Operation{
		OperationID: "update-transaction",
		Method:      http.MethodPut,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, transactionsKey, idKey),
		Summary:     "Update a Transaction",
		Description: "Update a specific Transaction",
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.WriteTransactionsPermission)},
		Tags:        []string{transactionsKey},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.updateTransactionHandler)

	huma.Register(api, huma.Operation{
		OperationID: "delete-transaction",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, transactionsKey, idKey),
		Summary:     "Delete a Transaction",
		Description: "Delete a specific Transaction",
		Tags:        []string{transactionsKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.WriteTransactionsPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
	}, app.deleteTransactionHandler)
}

func (app *Application) registerSearch(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "search-books",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, searchKey, booksKey),
		Summary:     "Search Books",
		Description: "Search books based on specific parameters",
		Tags:        []string{searchKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadBooksPermission)},
		Security: []map[string][]string{
			{bearerSecKey: {}},
			{basicAuthKey: {}},
		},
		Parameters: []*huma.Param{
			{
				Name:   query.MinPagesKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   query.MaxPagesKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   query.MinEditionKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   query.MaxEditionKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   query.MinPublishedAtKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.MaxPublishedAtKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.TitleKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   query.ISBNKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   query.AuthorsKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, reflect.SliceOf(typeString)),
			},
			{
				Name:   query.PublishersKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, reflect.SliceOf(typeString)),
			},
			{
				Name:   query.GenresKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, reflect.SliceOf(typeString)),
			},
			{
				Name:   query.MinCopiesKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   query.MaxCopiesKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   query.MinBorrowedCopiesKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   query.MaxBorrowedCopiesKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
		},
	}, app.searchBookHandler)

	huma.Register(api, huma.Operation{
		OperationID: "search-patrons",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, searchKey, patronsKey),
		Summary:     "Search Patrons",
		Description: "Search patrons based on specific parameters",
		Tags:        []string{searchKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadPatronsPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
		Parameters: []*huma.Param{
			{
				Name:   query.NameKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   query.EmailKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   query.CategoryKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
		},
	}, app.searchPatronsHandler)

	huma.Register(api, huma.Operation{
		OperationID: "search-transactions",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, searchKey, transactionsKey),
		Summary:     "Search Transactions",
		Description: "Search transactions based on specific parameters",
		Tags:        []string{searchKey},
		Middlewares: huma.Middlewares{app.authenticate(api), app.requirePermission(api, auth.ReadTransactionsPermission)},
		Security: []map[string][]string{
			{basicAuthKey: {}},
		},
		Parameters: []*huma.Param{
			{
				Name:   query.PatronIDKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   query.BookIDKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   query.StatusKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   query.MinBorrowedAtKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.MaxBorrowedAtKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.MinDueDateKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.MaxDueDateKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.MinReturnedAtKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.MaxReturnedAtKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.MinCreatedAtKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   query.MaxCreatedAtKey,
				In:     query.Key,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
		},
	}, app.searchTransactionsHandler)
}

func (app *Application) registerToken(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "create-auth-token",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, tokensKey, authenticationKey),
		Summary:     "Create an auth token",
		Description: "Create a specific auth token",
		Tags:        []string{tokensKey},
	}, app.createAuthTokenHandler)
}
