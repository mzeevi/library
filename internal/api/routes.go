package api

import (
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"net/http"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

const (
	basePath        = ""
	booksKey        = "books"
	patronsKey      = "patrons"
	transactionsKey = "transactions"
	borrowKey       = "borrow"
	returnKey       = "return"
	searchKey       = "search"
	healthcheckKey  = "healthcheck"
	idKey           = "id"
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

	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(httplog.RequestLogger(app.logger))
	router.Use(middleware.Recoverer)

	api := humachi.New(router, conf)

	app.registerHealthcheck(api)
	app.registerBooks(api)
	app.registerPatrons(api)
	app.registerTransactions(api)
	app.registerSearch(api)

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
	}, app.getBookHandler)

	huma.Register(api, huma.Operation{
		OperationID: "get-books",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s", basePath, booksKey),
		Summary:     "Get Books",
		Description: "Get all Books",
		Tags:        []string{booksKey},
	}, app.getBooksHandler)

	huma.Register(api, huma.Operation{
		OperationID: "create-book",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s", basePath, booksKey),
		Summary:     "Create a Book",
		Description: "Create a specific Book",
		Tags:        []string{booksKey},
	}, app.createBookHandler)

	huma.Register(api, huma.Operation{
		OperationID: "update-book",
		Method:      http.MethodPut,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, booksKey, idKey),
		Summary:     "Update a Book",
		Description: "Update a specific Book",
		Tags:        []string{booksKey},
	}, app.updateBookHandler)

	huma.Register(api, huma.Operation{
		OperationID: "delete-book",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, booksKey, idKey),
		Summary:     "Delete a Book",
		Description: "Delete a specific Book",
		Tags:        []string{booksKey},
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
	}, app.getPatronHandler)

	huma.Register(api, huma.Operation{
		OperationID: "get-patrons",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s", basePath, patronsKey),
		Summary:     "Get Patrons",
		Description: "Get all Patrons",
		Tags:        []string{patronsKey},
	}, app.getPatronsHandler)

	huma.Register(api, huma.Operation{
		OperationID: "create-patron",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s", basePath, patronsKey),
		Summary:     "Create a Patron",
		Description: "Create a specific Patron",
		Tags:        []string{patronsKey},
	}, app.createPatronHandler)

	huma.Register(api, huma.Operation{
		OperationID: "update-patron",
		Method:      http.MethodPut,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, patronsKey, idKey),
		Summary:     "Update a Patron",
		Description: "Update a specific Patron",
		Tags:        []string{patronsKey},
	}, app.updatePatronHandler)

	huma.Register(api, huma.Operation{
		OperationID: "delete-patron",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, patronsKey, idKey),
		Summary:     "Delete a Patron",
		Description: "Delete a specific Patron",
		Tags:        []string{patronsKey},
	}, app.deletePatronHandler)
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
	}, app.getTransactionHandler)

	huma.Register(api, huma.Operation{
		OperationID: "get-transactions",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/%s", basePath, transactionsKey),
		Summary:     "Get Transactions",
		Description: "Get all Transactions",
		Tags:        []string{transactionsKey},
	}, app.getTransactionsHandler)

	huma.Register(api, huma.Operation{
		OperationID: "borrow-book-transaction",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, transactionsKey, borrowKey),
		Summary:     "Borrow Book",
		Description: "Borrow a Book",
		Tags:        []string{transactionsKey},
	}, app.borrowBookTransactionHandler)

	huma.Register(api, huma.Operation{
		OperationID: "return-book-transaction",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/%s/%s", basePath, transactionsKey, returnKey),
		Summary:     "Return Book",
		Description: "Return a Book",
		Tags:        []string{transactionsKey},
	}, app.returnBookTransactionHandler)

	huma.Register(api, huma.Operation{
		OperationID: "update-transaction",
		Method:      http.MethodPut,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, transactionsKey, idKey),
		Summary:     "Update a Transaction",
		Description: "Update a specific Transaction",
		Tags:        []string{transactionsKey},
	}, app.updateTransactionHandler)

	huma.Register(api, huma.Operation{
		OperationID: "delete-transaction",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/%s/{%s}", basePath, transactionsKey, idKey),
		Summary:     "Delete a Transaction",
		Description: "Delete a specific Transaction",
		Tags:        []string{transactionsKey},
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
		Parameters: []*huma.Param{
			{
				Name:   minPagesQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   maxPagesQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   minEditionQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   maxEditionQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   minPublishedAtQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   maxPublishedAtQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   titleQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   isbnQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   authorsQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, reflect.SliceOf(typeString)),
			},
			{
				Name:   publishersQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, reflect.SliceOf(typeString)),
			},
			{
				Name:   genresQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, reflect.SliceOf(typeString)),
			},
			{
				Name:   minCopiesQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   maxCopiesQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   minBorrowedCopiesQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeInt),
			},
			{
				Name:   maxBorrowedCopiesQuery,
				In:     queryKey,
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
		Parameters: []*huma.Param{
			{
				Name:   nameQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   emailQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   categoryQuery,
				In:     queryKey,
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
		Parameters: []*huma.Param{
			{
				Name:   patronIDQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   bookIDQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   statusQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeString),
			},
			{
				Name:   minBorrowedAtQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   maxBorrowedAtQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   minDueDateQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   maxDueDateQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   minReturnedAtQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   maxReturnedAtQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   minCreatedAtQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
			{
				Name:   maxCreatedAtQuery,
				In:     queryKey,
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, typeTime),
			},
		},
	}, app.searchTransactionsHandler)
}
